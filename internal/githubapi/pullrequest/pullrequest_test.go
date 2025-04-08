package pullrequest_test

import (
	"os"
	"testing"

	"log/slog"
	"net/http"

	"github.com/google/go-github/v70/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/pullrequest"
	"github.com/stretchr/testify/assert"
)

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
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

func TestGetPRSHA_Failure(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "failed to get PR", http.StatusInternalServerError)
			}),
		),
	))

	sha, err := pullrequest.GetPRSHA("owner", "repo", 1, mockClient, logger)
	assert.Error(t, err)
	assert.Empty(t, sha)
}

func TestRemoveLabel_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK) // Ensure a valid HTTP 200 response
			}),
		),
	))

	err := pullrequest.RemoveLabel("owner", "repo", 1, "test-label", mockClient, logger)
	assert.NoError(t, err)
}

func TestRemoveLabel_Failure(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "failed to remove label", http.StatusInternalServerError)
			}),
		),
	))

	err := pullrequest.RemoveLabel("owner", "repo", 1, "test-label", mockClient, logger)
	assert.Error(t, err)
}

func TestAddComment_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			github.IssueComment{ID: github.Int64(1)},
		),
	))

	err := pullrequest.AddComment("owner", "repo", 1, "test comment", mockClient, logger)
	assert.NoError(t, err)
}

func TestAddComment_Failure(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "failed to add comment", http.StatusInternalServerError)
			}),
		),
	))

	err := pullrequest.AddComment("owner", "repo", 1, "test comment", mockClient, logger)
	assert.Error(t, err)
}
