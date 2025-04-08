package pullrequest

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v70/github"
)

// Fetches PR SHA
func GetPRSHA(owner, repo string, prNumber int, client *github.Client, logger *slog.Logger) (string, error) {
	pr, _, err := client.PullRequests.Get(context.Background(), owner, repo, prNumber)

	if err != nil {
		logger.Error("failed to get pull request", slog.String("error", err.Error()))
		return "", err
	}

	return *pr.Head.SHA, nil
}

func RemoveLabel(owner, repo string, prNumber int, label string, client *github.Client, logger *slog.Logger) error {
	ctx := context.Background()
	_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, label)
	if err != nil {
		if logger != nil {
			logger.Warn("Failed to remove label", slog.String("label", label), slog.String("error", err.Error()))
		}
		return err
	}
	if logger != nil {
		logger.Info("Removed label", slog.String("label", label))
	}

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
