package pullrequest_test

import (
	"testing"

	"log/slog"

	"github.com/google/go-github/v69/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/pullrequest"
	"github.com/stretchr/testify/assert"
)

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestGetPRSHA_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{
				Head: &github.PullRequestBranch{SHA: github.Ptr("test-sha")},
			},
		),
	))

	sha, err := pullrequest.GetPRSHA("owner", "repo", 1, mockClient, logger)
	assert.NoError(t, err)
	assert.Equal(t, "test-sha", sha)
}
