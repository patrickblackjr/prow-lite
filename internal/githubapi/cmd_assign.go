package githubapi

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/go-github/v71/github"
)

func assignUsers(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 || len(parts) > 4 {
		logger.Warn("invalid number of users to assign", slog.String("command", command))
		return
	}
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()

	for _, user := range parts[1:] {
		user = strings.TrimPrefix(user, "@")
		if _, _, err := client.Issues.AddAssignees(ctx, owner, repo, prNumber, []string{user}); err != nil {
			logger.Error("failed to assign user", slog.String("user", user), slog.String("error", err.Error()))
			continue
		}
		logger.Info("assigned user", slog.String("user", user))
	}
}

func unassignUsers(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 || len(parts) > 4 {
		logger.Warn("invalid number of users to unassign", slog.String("command", command))
		return
	}
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()

	for _, user := range parts[1:] {
		user = strings.TrimPrefix(user, "@")
		if _, _, err := client.Issues.RemoveAssignees(ctx, owner, repo, prNumber, []string{user}); err != nil {
			logger.Error("failed to unassign user", slog.String("user", user), slog.String("error", err.Error()))
			continue
		}
		logger.Info("unassigned user", slog.String("user", user))
	}
}
