package pullrequest

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v71/github"
)

func GetPRSHA(owner, repo string, prNumber int, client *github.Client, logger *slog.Logger) (string, error) {
	pr, _, err := client.PullRequests.Get(context.Background(), owner, repo, prNumber)
	if err != nil {
		logger.Error("failed to get pull request", slog.String("error", err.Error()))
		return "", err
	}
	return pr.GetHead().GetSHA(), nil
}

func RemoveLabel(owner, repo string, prNumber int, label string, client *github.Client, logger *slog.Logger) error {
	_, err := client.Issues.RemoveLabelForIssue(context.Background(), owner, repo, prNumber, label)
	if err != nil {
		logger.Warn("failed to remove label", slog.String("label", label), slog.String("error", err.Error()))
		return err
	}
	logger.Info("removed label", slog.String("label", label))
	return nil
}

func AddComment(owner, repo string, prNumber int, commentText string, client *github.Client, logger *slog.Logger) error {
	_, _, err := client.Issues.CreateComment(context.Background(), owner, repo, prNumber, &github.IssueComment{
		Body: github.Ptr(commentText),
	})
	if err != nil {
		logger.Error("failed to add comment", slog.String("comment_text", commentText))
		return err
	}
	logger.Info("added comment to PR", slog.String("comment_text", commentText))
	return nil
}
