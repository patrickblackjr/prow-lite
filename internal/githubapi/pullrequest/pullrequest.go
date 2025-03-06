package pullrequest

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v68/github"
)

// Fetches PR SHA
func GetPRSHA(owner, repo string, prNumber int, client *github.Client, logger *slog.Logger) string {
	ctx := context.Background()
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		logger.Error("Failed to retrieve PR SHA", slog.String("error", err.Error()))
		return ""
	}
	return pr.GetHead().GetSHA()
}

func RemoveLabel(owner, repo string, prNumber int, label string, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()
	_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, label)
	if err != nil {
		logger.Warn("Failed to remove label", slog.String("label", label), slog.String("error", err.Error()))
		return
	}
	logger.Info("Removed label", slog.String("label", label))
}

func AddComment(owner, repo string, prNumber int, commentText string, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, &github.IssueComment{Body: github.Ptr(commentText)})
	if err != nil {
		logger.Error("failed to add comment", slog.String("comment_text", commentText))
		return
	}
	logger.Info("added comment to PR", slog.String("comment_text", commentText))
}
