package githubapi

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v71/github"
)

func lgtm(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()
	user := event.GetComment().GetUser().GetLogin()

	if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{"lgtm"}); err != nil {
		logger.Error("failed to add lgtm label", slog.String("user", user), slog.String("error", err.Error()))
		return
	}
	logger.Info("added lgtm label", slog.String("user", user))

	if _, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, "do-not-merge"); err != nil {
		logger.Warn("failed to remove do-not-merge label", slog.String("user", user), slog.String("error", err.Error()))
	}

	if err := updateApprovalCheckRun(ctx, client, owner, repo, prNumber, "success", "Approved and ready for merge", "The pull request has been approved."); err != nil {
		logger.Error("failed to update approval check run", slog.String("user", user), slog.String("error", err.Error()))
		return
	}
	logger.Info("approval granted", slog.String("user", user))
}

func unlgtm(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()
	user := event.GetComment().GetUser().GetLogin()

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

	if err := updateApprovalCheckRun(ctx, client, owner, repo, prNumber, "neutral", "Approval revoked", "This PR is no longer approved."); err != nil {
		logger.Error("failed to update approval check run", slog.String("user", user), slog.String("error", err.Error()))
		return
	}
	logger.Info("approval revoked", slog.String("user", user))
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
