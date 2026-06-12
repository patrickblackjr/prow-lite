package githubapi

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v71/github"
)

var lgtmTriggers = []string{"/lgtm", "/approve"}
var unlgtmTriggers = []string{"/unlgtm", "/unapprove", "/remove-lgtm", "/remove-approve", "/remove-approval"}

// ApprovalResetComment is posted when a PR is reopened to signal that previous approvals are void.
// countApprovals uses it as a marker to discard approvals granted before the reopen.
const ApprovalResetComment = "Approval has been reset since this PR was reopened."

func lgtm(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger, minApprovals int) {
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()
	user := event.GetComment().GetUser().GetLogin()

	count := 0
	if minApprovals > 0 {
		var err error
		count, err = countApprovals(ctx, client, owner, repo, prNumber)
		if err != nil {
			logger.Error("failed to count approvals", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
	}
	logger.Info("counted approvals", slog.String("user", user), slog.Int("count", count), slog.Int("required", minApprovals))

	if count >= minApprovals {
		if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{"lgtm"}); err != nil {
			logger.Error("failed to add lgtm label", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
		if _, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, "do-not-merge"); err != nil {
			logger.Warn("failed to remove do-not-merge label", slog.String("user", user), slog.String("error", err.Error()))
		}
		summary := approvalSummary(count, minApprovals)
		if err := updateApprovalCheckRun(ctx, client, owner, repo, prNumber, "success", "Approved and ready for merge", summary); err != nil {
			logger.Error("failed to update approval check run", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
		logger.Info("approval granted", slog.String("user", user))
	} else {
		summary := approvalSummary(count, minApprovals)
		if err := updateApprovalCheckRun(ctx, client, owner, repo, prNumber, "neutral", "Approval needed", summary); err != nil {
			logger.Error("failed to update approval check run", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
		logger.Info("approval recorded, waiting for more", slog.String("user", user), slog.Int("count", count), slog.Int("required", minApprovals))
	}
}

func unlgtm(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger, minApprovals int) {
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()
	user := event.GetComment().GetUser().GetLogin()

	count := 0
	if minApprovals > 0 {
		var err error
		count, err = countApprovals(ctx, client, owner, repo, prNumber)
		if err != nil {
			logger.Error("failed to count approvals", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
	}
	logger.Info("counted approvals after revocation", slog.String("user", user), slog.Int("count", count), slog.Int("required", minApprovals))

	if count >= minApprovals {
		if err := updateApprovalCheckRun(ctx, client, owner, repo, prNumber, "success", "Approved and ready for merge", approvalSummary(count, minApprovals)); err != nil {
			logger.Error("failed to update approval check run", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
		logger.Info("approval revoked but still approved by others", slog.String("user", user))
	} else {
		if _, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, "lgtm"); err != nil {
			logger.Warn("failed to remove 'lgtm' label", slog.String("user", user), slog.String("error", err.Error()))
		} else {
			logger.Info("removed 'lgtm' label", slog.String("user", user))
		}
		if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{"do-not-merge"}); err != nil {
			logger.Warn("failed to add 'do-not-merge' label", slog.String("user", user), slog.String("error", err.Error()))
		} else {
			logger.Info("added 'do-not-merge' label", slog.String("user", user))
		}
		if err := updateApprovalCheckRun(ctx, client, owner, repo, prNumber, "neutral", "Approval revoked", approvalSummary(count, minApprovals)); err != nil {
			logger.Error("failed to update approval check run", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
		logger.Info("approval revoked", slog.String("user", user))
	}
}

func approvalSummary(count, minApprovals int) string {
	if minApprovals == 0 {
		return "Auto-approved: no approvals required."
	}
	return fmt.Sprintf("%d/%d approvals received.", count, minApprovals)
}

// countApprovals scans PR comments and returns the number of unique users with an active approval.
// An approval is active when a user's most recent lgtm/approve command has not been followed
// by an unlgtm/unapprove command in the same PR.
func countApprovals(ctx context.Context, client *github.Client, owner, repo string, prNumber int) (int, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	approved := make(map[string]bool)
	for {
		comments, resp, err := client.Issues.ListComments(ctx, owner, repo, prNumber, opts)
		if err != nil {
			return 0, fmt.Errorf("list comments: %w", err)
		}
		for _, c := range comments {
			if c.GetBody() == ApprovalResetComment {
				clear(approved)
				continue
			}
			user := c.GetUser().GetLogin()
			for _, line := range strings.Split(c.GetBody(), "\n") {
				line = strings.TrimSpace(line)
				for _, t := range lgtmTriggers {
					if strings.HasPrefix(line, t) {
						approved[user] = true
					}
				}
				for _, t := range unlgtmTriggers {
					if strings.HasPrefix(line, t) {
						delete(approved, user)
					}
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return len(approved), nil
}

// updateApprovalCheckRun creates an in-progress LGTM check run then immediately completes it.
// GitHub requires a check run to transition through in_progress before completing.
func updateApprovalCheckRun(ctx context.Context, client *github.Client, owner, repo string, prNumber int, conclusion, title, summary string) error {
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return fmt.Errorf("get pull request: %w", err)
	}

	checkRun, _, err := client.Checks.CreateCheckRun(ctx, owner, repo, github.CreateCheckRunOptions{
		Name:    "LGTM",
		HeadSHA: pr.GetHead().GetSHA(),
		Status:  github.Ptr("in_progress"),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	})
	if err != nil {
		return fmt.Errorf("create check run: %w", err)
	}

	_, _, err = client.Checks.UpdateCheckRun(ctx, owner, repo, checkRun.GetID(), github.UpdateCheckRunOptions{
		Name:       "LGTM",
		Status:     github.Ptr("completed"),
		Conclusion: github.Ptr(conclusion),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	})
	if err != nil {
		return fmt.Errorf("update check run: %w", err)
	}

	return nil
}
