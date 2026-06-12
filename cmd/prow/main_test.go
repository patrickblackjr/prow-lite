package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func captureExit(t *testing.T) *int {
	t.Helper()
	code := -1
	osExit = func(c int) { code = c }
	t.Cleanup(func() { osExit = os.Exit })
	return &code
}

func noopServer(t *testing.T) {
	t.Helper()
	runServer = func(r *gin.Engine, addr string) error { return nil }
	t.Cleanup(func() { runServer = func(r *gin.Engine, addr string) error { return r.Run(addr) } })
}

func writeLabelSyncConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	labelsPath := filepath.Join(dir, "labels.yml")
	require.NoError(t, os.WriteFile(labelsPath, []byte("categories: []\n"), 0o644))
	prowCfg := "github:\n  app_id: 1\n  installation_id: 1\nfeatures:\n  label_sync:\n    path: " + labelsPath + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)
}

func TestMain_CLIError(t *testing.T) {
	// Missing required --mode flag causes cmd.Run to return an error.
	oldArgs := os.Args
	os.Args = []string{"prow", "run"} // missing required --mode
	t.Cleanup(func() { os.Args = oldArgs })
	t.Chdir(t.TempDir())
	main()
}

func TestMain_ActionExitsWithCode2(t *testing.T) {
	code := captureExit(t)
	oldArgs := os.Args
	os.Args = []string{"prow", "run", "--mode", "ci", "--plugin", "event", "--event", `{"action":"x"}`}
	t.Cleanup(func() { os.Args = oldArgs })
	t.Chdir(t.TempDir()) // no config → NewGithubClient fails → osExit(2)
	main()
	assert.Equal(t, 2, *code)
}

func TestSetupRouter(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient())
	r := setupRouter(c, discardLogger(), githubapi.ProcessComment, githubapi.NewPREventHandler(1))
	require.NotNil(t, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRunAction_Standalone(t *testing.T) {
	noopServer(t)
	runAction(context.Background(), "standalone", "", "", github.NewClient(nil), discardLogger())
}

func TestRunAction_Standalone_ServerError(t *testing.T) {
	runServer = func(r *gin.Engine, addr string) error { return assert.AnError }
	t.Cleanup(func() { runServer = func(r *gin.Engine, addr string) error { return r.Run(addr) } })
	runAction(context.Background(), "standalone", "", "", github.NewClient(nil), discardLogger())
}

func TestRunAction_CI_Event_EmptyEvent(t *testing.T) {
	code := captureExit(t)
	runAction(context.Background(), "ci", "event", "", github.NewClient(nil), discardLogger())
	assert.Equal(t, 1, *code)
}

func TestRunAction_CI_Event_InvalidJSON(t *testing.T) {
	code := captureExit(t)
	runAction(context.Background(), "ci", "event", "not-json", github.NewClient(nil), discardLogger())
	assert.Equal(t, 1, *code)
}

func TestRunAction_CI_Event_NoAction(t *testing.T) {
	code := captureExit(t)
	runAction(context.Background(), "ci", "event", `{"foo":"bar"}`, github.NewClient(nil), discardLogger())
	assert.Equal(t, 1, *code)
}

func TestRunAction_CI_Event_Success(t *testing.T) {
	t.Chdir(t.TempDir())
	runAction(context.Background(), "ci", "event", `{"action":"unknown_event"}`, github.NewClient(nil), discardLogger())
}

func TestRunAction_CI_LabelSync_NoConfig(t *testing.T) {
	code := captureExit(t)
	t.Chdir(t.TempDir())
	runAction(context.Background(), "ci", "labelsync", "", github.NewClient(nil), discardLogger())
	assert.Equal(t, 1, *code)
}

func TestRunAction_CI_LabelDashSync_NoConfig(t *testing.T) {
	code := captureExit(t)
	t.Chdir(t.TempDir())
	runAction(context.Background(), "ci", "label-sync", "", github.NewClient(nil), discardLogger())
	assert.Equal(t, 1, *code)
}

func TestRunAction_CI_LabelSync_RunFails(t *testing.T) {
	code := captureExit(t)
	writeLabelSyncConfig(t)
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(mock.GetInstallationRepositories,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	runAction(context.Background(), "ci", "labelsync", "", c, discardLogger())
	assert.Equal(t, 1, *code)
}

func TestRunAction_CI_LabelSync_Success(t *testing.T) {
	writeLabelSyncConfig(t)
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetInstallationRepositories, github.ListRepositories{
			Repositories: []*github.Repository{
				{
					Name:     github.Ptr("repo"),
					FullName: github.Ptr("owner/repo"),
					Owner:    &github.User{Login: github.Ptr("owner")},
					Archived: github.Ptr(false),
					Fork:     github.Ptr(false),
				},
			},
		}),
		// No labels to sync (empty config), no GetLabel calls needed
	))
	runAction(context.Background(), "ci", "labelsync", "", c, discardLogger())
}

func TestRunAction_CI_UnknownPlugin(t *testing.T) {
	code := captureExit(t)
	runAction(context.Background(), "ci", "unknown-plugin", "", github.NewClient(nil), discardLogger())
	assert.Equal(t, 1, *code)
}

func TestRunAction_UnknownMode(t *testing.T) {
	runAction(context.Background(), "unknown-mode", "", "", github.NewClient(nil), discardLogger())
}

func TestRunAction_CI_Event_MinApprovals(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	prowCfg := "github:\n  app_id: 1\n  installation_id: 1\nfeatures:\n  lgtm:\n    min_approvals: 2\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)
	runAction(context.Background(), "ci", "event", `{"action":"unknown_event"}`, github.NewClient(nil), discardLogger())
}

func TestRunAction_CI_Event_ZeroMinApprovals(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	prowCfg := "github:\n  app_id: 1\n  installation_id: 1\nfeatures:\n  lgtm:\n    min_approvals: 0\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)
	runAction(context.Background(), "ci", "event", `{"action":"unknown_event"}`, github.NewClient(nil), discardLogger())
}
