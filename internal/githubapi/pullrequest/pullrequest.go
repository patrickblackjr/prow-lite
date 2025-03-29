package pullrequest

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v69/github"
)

// Fetches PR SHA
func GetPRSHA(owner, repo string, prNumber int, client *github.Client, logger *slog.Logger) (string, error) {
	ctx := context.Background()
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		logger.Error("failed to retrieve PR SHA", slog.String("error", err.Error()))
		return "", err
	}
	if pr != nil {
		return pr.GetHead().GetSHA(), nil
	}
	return "", nil
}

func RemoveLabel(owner, repo string, prNumber int, label string, client *github.Client, logger *slog.Logger) error {
	ctx := context.Background()
	_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, label)
	if err != nil {
		logger.Warn("Failed to remove label", slog.String("label", label), slog.String("error", err.Error()))
		return err
	}
	logger.Info("Removed label", slog.String("label", label))
	return nil
}

func AddComment(owner, repo string, prNumber int, commentText string, client *github.Client, logger *slog.Logger) error {
	ctx := context.Background()
	comment := &github.IssueComment{Body: github.Ptr(commentText)}

	_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	if err != nil {
		logger.Error("failed to add comment", slog.String("comment_text", commentText))
		return err
	}
	logger.Info("added comment to PR", slog.String("comment_text", commentText))
	return nil
}
