package plugins

import (
	"context"
	"strings"

	"log/slog"

	"github.com/google/go-github/v68/github"
)

func ProcessComment(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	comment := *event.Comment.Body
	commands := strings.Split(comment, "\n")
	for _, command := range commands {
		if strings.HasPrefix(command, "/assign") {
			assignUsers(command, event, client, logger)
		}
		if strings.HasPrefix(command, "/unassign") {
			unassignUsers(command, event, client, logger)
		}
		if strings.HasPrefix(command, "/lgtm") || strings.HasPrefix(command, "/approve") {
			lgtm(event, client, logger)
		}
	}
}

func assignUsers(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	// Extract usernames from the command
	parts := strings.Fields(command)
	if len(parts) < 2 || len(parts) > 4 {
		logger.Warn("invalid number of users to assign", slog.String("command", command))
		return
	}
	users := parts[1:]
	ctx := context.Background()
	for _, user := range users {
		user = strings.TrimPrefix(user, "@")
		_, _, err := client.Issues.AddAssignees(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Issue.Number, []string{user})
		if err != nil {
			logger.Error("failed to assign user", slog.String("user", user), slog.String("error", err.Error()))
			continue
		}
		logger.Info("assigning user", slog.String("user", user))
	}
}

func unassignUsers(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	// Extract usernames from the command
	parts := strings.Fields(command)
	if len(parts) < 2 || len(parts) > 4 {
		logger.Warn("invalid number of users to unassign", slog.String("command", command))
		return
	}
	users := parts[1:]
	ctx := context.Background()
	for _, user := range users {
		user = strings.TrimPrefix(user, "@")
		_, _, err := client.Issues.RemoveAssignees(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Issue.Number, []string{user})
		if err != nil {
			logger.Error("failed to unassign user", slog.String("user", user), slog.String("error", err.Error()))
			continue
		}
		logger.Info("unassigning user", slog.String("user", user))
	}
}

func lgtm(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := *event.Issue.Number
	user := *event.Comment.User.Login

	// Check if the user is trying to approve their own pull request
	if *event.Issue.User.Login == user {
		logger.Warn("user cannot approve their own pull request", slog.String("user", user))
		return
	}

	// Ensure the "lgtm" label exists
	_, _, err := client.Issues.GetLabel(ctx, owner, repo, "lgtm")
	if err != nil {
		_, _, err = client.Issues.CreateLabel(ctx, owner, repo, &github.Label{
			Name:        github.Ptr("lgtm"),
			Color:       github.Ptr("0e8a16"),
			Description: github.Ptr("Approved by reviewers"),
		})
		if err != nil {
			logger.Error("failed to create lgtm label", slog.String("user", user), slog.String("error", err.Error()))
			return
		}
		logger.Info("created lgtm label", slog.String("user", user))
	}

	// Add the "lgtm" label to the pull request
	_, _, err = client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{"lgtm"})
	if err != nil {
		logger.Error("failed to add lgtm label", slog.String("user", user), slog.String("error", err.Error()))
		return
	}
	logger.Info("added lgtm label", slog.String("user", user))

	// Retrieve the pull request to get the SHA
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		logger.Error("failed to get pull request", slog.String("user", user), slog.String("error", err.Error()))
		return
	}

	// Create a check run to indicate the approval
	checkRun := &github.CreateCheckRunOptions{
		Name:    "LGTM",
		HeadSHA: pr.GetHead().GetSHA(),
		Status:  github.Ptr("in_progress"),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr("LGTM"),
			Summary: github.Ptr("The pull request is awaiting approval."),
		},
	}
	checkRunResult, _, err := client.Checks.CreateCheckRun(ctx, owner, repo, *checkRun)
	if err != nil {
		logger.Error("failed to create check run", slog.String("user", user), slog.String("error", err.Error()))
		return
	}
	logger.Info("created check run", slog.String("user", user))

	// Remove the "do-not-merge" label
	_, err = client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, "do-not-merge")
	if err != nil {
		logger.Error("failed to remove do-not-merge label", slog.String("user", user), slog.String("error", err.Error()))
		return
	}
	logger.Info("removed do-not-merge label", slog.String("user", user))

	// Update the check run to indicate the approval
	updateCheckRun := &github.UpdateCheckRunOptions{
		Status:     github.Ptr("completed"),
		Conclusion: github.Ptr("success"),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr("LGTM"),
			Summary: github.Ptr("The pull request has been approved."),
		},
	}
	_, _, err = client.Checks.UpdateCheckRun(ctx, owner, repo, checkRunResult.GetID(), *updateCheckRun)
	if err != nil {
		logger.Error("failed to update check run", slog.String("user", user), slog.String("error", err.Error()))
		return
	}
	logger.Info("updated check run", slog.String("user", user))
}
