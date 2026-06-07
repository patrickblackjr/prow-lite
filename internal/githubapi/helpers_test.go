package githubapi

import (
	"io"
	"log/slog"

	"github.com/google/go-github/v71/github"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func makeIssueCommentEvent(owner, repo string, prNumber int, commentBody string) *github.IssueCommentEvent {
	return &github.IssueCommentEvent{
		Repo: &github.Repository{
			Name:  github.Ptr(repo),
			Owner: &github.User{Login: github.Ptr(owner)},
		},
		Issue:   &github.Issue{Number: github.Ptr(prNumber)},
		Comment: &github.IssueComment{Body: github.Ptr(commentBody), User: &github.User{Login: github.Ptr("commenter")}},
	}
}
