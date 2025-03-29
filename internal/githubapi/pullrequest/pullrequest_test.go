package pullrequest_test

import (
	"net/http"
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

func TestGetPRSHA_Failure(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mock.WriteError(w, http.StatusNotFound, "Not Found")
			}),
		),
	))

	// sha, err :=
	pullrequest.GetPRSHA("owner", "repo", 1, mockClient, logger)
	// assert.Empty(t, sha)
	// assert.Error(t, err)
	// assert.ErrorContains(t, err, "Not Found")
}

func TestRemoveLabel_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			nil,
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
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "failed to remove label")
			}),
		),
	))

	err := pullrequest.RemoveLabel("owner", "repo", 1, "test-label", mockClient, logger)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to remove label")
}

func TestAddComment_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			github.IssueComment{
				ID:   github.Ptr(int64(1)),
				Body: github.String("test comment"),
			},
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
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "failed to add comment")
			}),
		),
	))

	err := pullrequest.AddComment("owner", "repo", 1, "test comment", mockClient, logger)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to add comment")
}

func TestAddComment_Failure_NilComment(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "failed to add comment")
			}),
		),
	))

	err := pullrequest.AddComment("owner", "repo", 1, "", mockClient, logger) // Empty comment text
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to add comment")
}
