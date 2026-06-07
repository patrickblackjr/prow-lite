package githubapi

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/go-github/v71/github"
)

func label(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		logger.Warn("missing label name", slog.String("command", command))
		return
	}
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()
	labelName := parts[1]

	if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{labelName}); err != nil {
		logger.Error("failed to add label", slog.String("label", labelName), slog.String("error", err.Error()))
		return
	}
	logger.Info("added label", slog.String("label", labelName))
}

func removeLabel(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		logger.Warn("missing label name", slog.String("command", command))
		return
	}
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()
	labelName := parts[1]

	if _, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, labelName); err != nil {
		logger.Error("failed to remove label", slog.String("label", labelName), slog.String("error", err.Error()))
		return
	}
	logger.Info("removed label", slog.String("label", labelName))
}
