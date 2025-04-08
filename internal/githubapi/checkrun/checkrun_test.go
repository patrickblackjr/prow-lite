package checkrun_test

import (
	"net/http"
	"os"
	"testing"

	"log/slog"

	"github.com/google/go-github/v70/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/checkrun"
	"github.com/stretchr/testify/assert"
)

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestCreateCheckRun_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{
				ID: github.Ptr(int64(1)),
			},
		),
	))
	checkRun, err := checkrun.CreateCheckRun("owner", "repo", "sha", "success", "Test", mockClient, logger)
	assert.NoError(t, err)
	assert.NotNil(t, checkRun)
	assert.Equal(t, int64(1), checkRun.GetID())
	assert.Nil(t, err)
}

func TestCreateCheckRun_Failure(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposCheckRunsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "failed to create check run")
			}),
		),
	))
	checkRun, err := checkrun.CreateCheckRun("owner", "repo", "sha", "failure", "Test", mockClient, logger)
	assert.Error(t, err)
	assert.Nil(t, checkRun)
	assert.ErrorContains(t, err, "failed to create check run")
}
