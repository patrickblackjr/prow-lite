package githubapi

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/google/go-github/v71/github"
)

func label(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		logger.Warn("missing label name", slog.String("command", command))
		return
	}

	var labelName string
	if len(parts) >= 3 {
		// /label category name → category/name
		labelName = parts[1] + "/" + parts[2]
	} else {
		// /label category/name or /label category:name
		labelName = strings.ReplaceAll(parts[1], ":", "/")
	}

	addLabel(labelName, event, client, logger)
}

func applyCategoryLabel(command, category string, validLabels []string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		logger.Warn("missing label name", slog.String("command", command))
		return
	}
	labelSuffix := parts[1]

	if !slices.Contains(validLabels, labelSuffix) {
		logger.Warn("unknown label for category", slog.String("category", category), slog.String("label", labelSuffix))
		return
	}

	addLabel(category+"/"+labelSuffix, event, client, logger)
}

func addLabel(labelName string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()

	if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{labelName}); err != nil {
		logger.Error("failed to add label", slog.String("label", labelName), slog.String("error", err.Error()))
		return
	}
	logger.Info("added label", slog.String("label", labelName))
}

func unlabel(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		logger.Warn("missing label name", slog.String("command", command))
		return
	}

	var labelName string
	if len(parts) >= 3 {
		labelName = parts[1] + "/" + parts[2]
	} else {
		labelName = strings.ReplaceAll(parts[1], ":", "/")
	}

	deleteLabel(labelName, event, client, logger)
}

func removeCategoryLabel(command, category string, validLabels []string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		logger.Warn("missing label name", slog.String("command", command))
		return
	}
	labelSuffix := parts[1]

	if !slices.Contains(validLabels, labelSuffix) {
		logger.Warn("unknown label for category", slog.String("category", category), slog.String("label", labelSuffix))
		return
	}

	deleteLabel(category+"/"+labelSuffix, event, client, logger)
}

func deleteLabel(labelName string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	number := event.GetIssue().GetNumber()

	if _, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, labelName); err != nil {
		logger.Error("failed to remove label", slog.String("label", labelName), slog.String("error", err.Error()))
		return
	}
	logger.Info("removed label", slog.String("label", labelName))
}

// removeLabel is the legacy handler for /remove-label, kept for backwards compatibility.
func removeLabel(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	unlabel(command, event, client, logger)
}
