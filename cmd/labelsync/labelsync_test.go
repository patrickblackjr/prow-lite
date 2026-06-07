package labelsync

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func generateTestRSAKey(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// writeProwAndLabelConfig sets up a temp dir with prow-lite.yml pointing to a labels file.
func writeProwAndLabelConfig(t *testing.T, labelContent string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))

	labelsPath := filepath.Join(dir, "labels.yml")
	require.NoError(t, os.WriteFile(labelsPath, []byte(labelContent), 0o644))

	prowCfg := "github:\n  app_id: 1\n  installation_id: 1\nfeatures:\n  label_sync:\n    path: " + labelsPath + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)
}

// captureExit replaces osExit for the duration of the test and returns the captured code.
func captureExit(t *testing.T) *int {
	t.Helper()
	code := -1
	osExit = func(c int) { code = c }
	t.Cleanup(func() { osExit = os.Exit })
	return &code
}

//  GetLabelSyncConfig

func TestGetLabelSyncConfig_Success(t *testing.T) {
	writeProwAndLabelConfig(t, `
overwrite: true
dry_run: true
categories:
  - name: type
    category_color: "0075ca"
    labels:
      - name: bug
extra_labels:
  - name: wip
    color: "b60205"
`)
	lsc, err := GetLabelSyncConfig(discardLogger())
	require.NoError(t, err)
	assert.True(t, lsc.Overwrite)
	assert.True(t, lsc.DryRun)
	assert.Len(t, lsc.Categories, 1)
	assert.Len(t, lsc.ExtraLabels, 1)
}

func TestGetLabelSyncConfig_NoProwConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := GetLabelSyncConfig(discardLogger())
	assert.Error(t, err)
}

func TestGetLabelSyncConfig_LabelsFileNotFound(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	prowCfg := "github:\n  app_id: 1\n  installation_id: 1\nfeatures:\n  label_sync:\n    path: /nonexistent/labels.yml\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)

	_, err := GetLabelSyncConfig(discardLogger())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestGetLabelSyncConfig_InvalidYAML(t *testing.T) {
	writeProwAndLabelConfig(t, `{invalid:[`)
	_, err := GetLabelSyncConfig(discardLogger())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetLabelSyncConfig_MissingColor(t *testing.T) {
	writeProwAndLabelConfig(t, `
categories:
  - name: type
    labels:
      - name: bug
`)
	_, err := GetLabelSyncConfig(discardLogger())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "color")
}

func TestSyncer_buildLabels(t *testing.T) {
	desc := github.Ptr("Something broken")
	tests := []struct {
		name string
		cfg  *LabelSyncConfig
		want []github.Label
	}{
		{
			name: "category color used when label has none",
			cfg: &LabelSyncConfig{
				Categories: []Category{
					{Name: "type", CategoryColor: "0075ca", Labels: []CategoryLabel{{Name: "bug", Description: desc}}},
				},
			},
			want: []github.Label{
				{Name: github.Ptr("type/bug"), Color: github.Ptr("0075ca"), Description: desc},
			},
		},
		{
			name: "label color overrides category color",
			cfg: &LabelSyncConfig{
				Categories: []Category{
					{Name: "type", CategoryColor: "0075ca", Labels: []CategoryLabel{{Name: "feature", Color: "e4e669"}}},
				},
			},
			want: []github.Label{
				{Name: github.Ptr("type/feature"), Color: github.Ptr("e4e669")},
			},
		},
		{
			name: "extra labels appended after category labels",
			cfg: &LabelSyncConfig{
				Categories: []Category{
					{Name: "type", CategoryColor: "0075ca", Labels: []CategoryLabel{{Name: "bug"}}},
				},
				ExtraLabels: []ExtraLabel{{Name: "wip", Color: "b60205"}},
			},
			want: []github.Label{
				{Name: github.Ptr("type/bug"), Color: github.Ptr("0075ca")},
				{Name: github.Ptr("wip"), Color: github.Ptr("b60205")},
			},
		},
		{
			name: "empty config produces nil slice",
			cfg:  &LabelSyncConfig{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSyncer(tt.cfg, nil, slog.Default())
			assert.Equal(t, tt.want, s.buildLabels())
		})
	}
}

func TestLabelNeedsUpdate(t *testing.T) {
	tests := []struct {
		name     string
		existing *github.Label
		desired  *github.Label
		want     bool
	}{
		{"color changed",
			&github.Label{Color: github.Ptr("ff0000")},
			&github.Label{Color: github.Ptr("00ff00")}, true},
		{"color unchanged, no description",
			&github.Label{Color: github.Ptr("ff0000")},
			&github.Label{Color: github.Ptr("ff0000")}, false},
		{"description changed",
			&github.Label{Color: github.Ptr("ff0000"), Description: github.Ptr("old")},
			&github.Label{Color: github.Ptr("ff0000"), Description: github.Ptr("new")}, true},
		{"description added where none existed",
			&github.Label{Color: github.Ptr("ff0000")},
			&github.Label{Color: github.Ptr("ff0000"), Description: github.Ptr("new")}, true},
		{"desired description nil...existing not touched",
			&github.Label{Color: github.Ptr("ff0000"), Description: github.Ptr("keep")},
			&github.Label{Color: github.Ptr("ff0000")}, false},
		{"nothing changed",
			&github.Label{Color: github.Ptr("ff0000"), Description: github.Ptr("same")},
			&github.Label{Color: github.Ptr("ff0000"), Description: github.Ptr("same")}, false},
		{"color and description both changed",
			&github.Label{Color: github.Ptr("ff0000"), Description: github.Ptr("old")},
			&github.Label{Color: github.Ptr("00ff00"), Description: github.Ptr("new")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, labelNeedsUpdate(tt.existing, tt.desired))
		})
	}
}

//  listRepos

func installationRepo(name, owner string, archived, fork bool) *github.Repository {
	return &github.Repository{
		Name:     github.Ptr(name),
		FullName: github.Ptr(owner + "/" + name),
		Owner:    &github.User{Login: github.Ptr(owner)},
		Archived: github.Ptr(archived),
		Fork:     github.Ptr(fork),
	}
}

func TestListRepos_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetInstallationRepositories, github.ListRepositories{
			Repositories: []*github.Repository{
				installationRepo("active", "org", false, false),
				installationRepo("archived", "org", true, false),
				installationRepo("forked", "org", false, true),
			},
		}),
	))
	s := NewSyncer(&LabelSyncConfig{}, c, discardLogger())
	repos, err := s.listRepos(context.Background())
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "active", repos[0].GetName())
}

func TestListRepos_APIError(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetInstallationRepositories,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	s := NewSyncer(&LabelSyncConfig{}, c, discardLogger())
	_, err := s.listRepos(context.Background())
	assert.Error(t, err)
}

func TestListRepos_Empty(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetInstallationRepositories, github.ListRepositories{}),
	))
	s := NewSyncer(&LabelSyncConfig{}, c, discardLogger())
	_, err := s.listRepos(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no repositories")
}

func TestSyncLabel_CreateNew(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(mock.GetReposLabelsByOwnerByRepoByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusNotFound, "not found")
			}),
		),
		mock.WithRequestMatch(mock.PostReposLabelsByOwnerByRepo,
			github.Label{Name: github.Ptr("bug")}),
	))
	NewSyncer(&LabelSyncConfig{}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")})
}

func TestSyncLabel_CreateNew_DryRun(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(mock.GetReposLabelsByOwnerByRepoByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusNotFound, "not found")
			}),
		),
	))
	NewSyncer(&LabelSyncConfig{DryRun: true}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")})
}

func TestSyncLabel_CreateNew_Fails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(mock.GetReposLabelsByOwnerByRepoByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusNotFound, "not found")
			}),
		),
		mock.WithRequestMatchHandler(mock.PostReposLabelsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	NewSyncer(&LabelSyncConfig{}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")})
}

func TestSyncLabel_Exists_OverwriteDisabled(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")}),
	))
	NewSyncer(&LabelSyncConfig{Overwrite: false}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")})
}

func TestSyncLabel_Exists_NoUpdate(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")}),
	))
	NewSyncer(&LabelSyncConfig{Overwrite: true}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")})
}

func TestSyncLabel_Update_DryRun(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")}),
	))
	NewSyncer(&LabelSyncConfig{Overwrite: true, DryRun: true}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("00ff00")})
}

func TestSyncLabel_Update_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")}),
		mock.WithRequestMatch(mock.PatchReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("bug"), Color: github.Ptr("00ff00")}),
	))
	NewSyncer(&LabelSyncConfig{Overwrite: true}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("00ff00")})
}

func TestSyncLabel_Update_Fails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")}),
		mock.WithRequestMatchHandler(mock.PatchReposLabelsByOwnerByRepoByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	NewSyncer(&LabelSyncConfig{Overwrite: true}, c, discardLogger()).syncLabel(
		context.Background(), "owner", "repo", "owner/repo",
		github.Label{Name: github.Ptr("bug"), Color: github.Ptr("00ff00")})
}

func TestPruneRepo_DeletesUnknownLabels(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepo, []*github.Label{
			{Name: github.Ptr("bug")},
			{Name: github.Ptr("keep")},
		}),
		mock.WithRequestMatch(mock.DeleteReposLabelsByOwnerByRepoByName, nil),
	))
	NewSyncer(&LabelSyncConfig{}, c, discardLogger()).pruneRepo(
		context.Background(), "owner", "repo", "owner/repo",
		map[string]struct{}{"keep": {}})
}

func TestPruneRepo_DryRun(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepo, []*github.Label{
			{Name: github.Ptr("old")},
		}),
	))
	NewSyncer(&LabelSyncConfig{DryRun: true}, c, discardLogger()).pruneRepo(
		context.Background(), "owner", "repo", "owner/repo", map[string]struct{}{})
}

func TestPruneRepo_DeleteFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepo, []*github.Label{
			{Name: github.Ptr("old")},
		}),
		mock.WithRequestMatchHandler(mock.DeleteReposLabelsByOwnerByRepoByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	NewSyncer(&LabelSyncConfig{}, c, discardLogger()).pruneRepo(
		context.Background(), "owner", "repo", "owner/repo", map[string]struct{}{})
}

func TestPruneRepo_ListFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(mock.GetReposLabelsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	NewSyncer(&LabelSyncConfig{}, c, discardLogger()).pruneRepo(
		context.Background(), "owner", "repo", "owner/repo", map[string]struct{}{})
}

func TestSyncer_Run_ListReposFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(mock.GetInstallationRepositories,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	err := NewSyncer(&LabelSyncConfig{}, c, discardLogger()).Run(context.Background())
	assert.Error(t, err)
}

func TestSyncer_Run_NoPrune(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetInstallationRepositories, github.ListRepositories{
			Repositories: []*github.Repository{installationRepo("repo", "owner", false, false)},
		}),
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("type/bug"), Color: github.Ptr("ff0000")}),
	))
	cfg := &LabelSyncConfig{
		Categories: []Category{
			{Name: "type", CategoryColor: "ff0000", Labels: []CategoryLabel{{Name: "bug"}}},
		},
	}
	err := NewSyncer(cfg, c, discardLogger()).Run(context.Background())
	assert.NoError(t, err)
}

func TestSyncer_Run_WithPrune(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetInstallationRepositories, github.ListRepositories{
			Repositories: []*github.Repository{installationRepo("repo", "owner", false, false)},
		}),
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepoByName,
			github.Label{Name: github.Ptr("type/bug"), Color: github.Ptr("ff0000")}),
		mock.WithRequestMatch(mock.GetReposLabelsByOwnerByRepo, []*github.Label{
			{Name: github.Ptr("type/bug")},
		}),
	))
	cfg := &LabelSyncConfig{
		Prune: true,
		Categories: []Category{
			{Name: "type", CategoryColor: "ff0000", Labels: []CategoryLabel{{Name: "bug"}}},
		},
	}
	err := NewSyncer(cfg, c, discardLogger()).Run(context.Background())
	assert.NoError(t, err)
}

// cancelAfterFirstRT cancels ctx after the first successful HTTP round-trip.
// Used to let listRepos succeed but cause sem.Acquire to fail in syncLabels.
type cancelAfterFirstRT struct {
	delegate http.RoundTripper
	cancel   context.CancelFunc
	once     sync.Once
}

func (c *cancelAfterFirstRT) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.delegate.RoundTrip(req)
	c.once.Do(c.cancel)
	return resp, err
}

func TestSyncer_syncLabels_ContextCancelled(t *testing.T) {
	old := maxConcurrentRequests
	maxConcurrentRequests = 0 // force sem.Acquire to block
	t.Cleanup(func() { maxConcurrentRequests = old })

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	repos := []*github.Repository{installationRepo("repo", "org", false, false)}
	labels := []github.Label{{Name: github.Ptr("bug"), Color: github.Ptr("ff0000")}}
	s := NewSyncer(&LabelSyncConfig{}, github.NewClient(nil), discardLogger())
	err := s.syncLabels(ctx, repos, labels)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "acquire semaphore")
}

func TestSyncer_pruneLabels_ContextCancelled(t *testing.T) {
	old := maxConcurrentRequests
	maxConcurrentRequests = 0
	t.Cleanup(func() { maxConcurrentRequests = old })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repos := []*github.Repository{installationRepo("repo", "org", false, false)}
	s := NewSyncer(&LabelSyncConfig{}, github.NewClient(nil), discardLogger())
	err := s.pruneLabels(ctx, repos, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "acquire semaphore")
}

func TestSyncer_Run_SyncLabelsFails(t *testing.T) {
	old := maxConcurrentRequests
	maxConcurrentRequests = 0
	t.Cleanup(func() { maxConcurrentRequests = old })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	baseMock := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetInstallationRepositories, github.ListRepositories{
			Repositories: []*github.Repository{installationRepo("repo", "owner", false, false)},
		}),
	)
	transport := &cancelAfterFirstRT{delegate: baseMock.Transport, cancel: cancel}
	c := github.NewClient(&http.Client{Transport: transport})

	cfg := &LabelSyncConfig{
		Categories: []Category{
			{Name: "type", CategoryColor: "ff0000", Labels: []CategoryLabel{{Name: "bug"}}},
		},
	}
	err := NewSyncer(cfg, c, discardLogger()).Run(ctx)
	assert.Error(t, err)
}

func TestPruneRepo_MultiPage(t *testing.T) {
	// Simulate pagination: the mock returns Link header pointing to page 2.
	page := 0
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(mock.GetReposLabelsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				page++
				if page == 1 {
					// Return first page with a Link header for page 2.
					w.Header().Set("Link",
						`<https://api.github.com/repos/owner/repo/labels?page=2>; rel="next"`)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[{"name":"keep"}]`))
			}),
		),
	))
	keepSet := map[string]struct{}{"keep": {}}
	NewSyncer(&LabelSyncConfig{}, c, discardLogger()).pruneRepo(
		context.Background(), "owner", "repo", "owner/repo", keepSet)
	assert.Equal(t, 2, page)
}

// fakeClientProvider wraps a *github.Client to satisfy githubClientProvider.
type fakeClientProvider struct{ c *github.Client }

func (f *fakeClientProvider) GetClient() *github.Client { return f.c }

func TestLabelSync_Success(t *testing.T) {
	old := newGithubClientFunc
	t.Cleanup(func() { newGithubClientFunc = old })

	mockC := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetInstallationRepositories, github.ListRepositories{
			Repositories: []*github.Repository{installationRepo("repo", "owner", false, false)},
		}),
		// No labels to sync (empty config)
	))
	newGithubClientFunc = func(logger *slog.Logger) (githubClientProvider, error) {
		return &fakeClientProvider{c: mockC}, nil
	}

	writeProwAndLabelConfig(t, "categories: []\n")
	LabelSync() // should complete without osExit
}

func TestLabelSync_NoConfig(t *testing.T) {
	code := captureExit(t)
	t.Chdir(t.TempDir()) // no prow-lite.yml
	LabelSync()
	assert.Equal(t, 2, *code)
}

func TestLabelSync_NoLabelSyncConfig(t *testing.T) {
	code := captureExit(t)
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))

	// generate a valid RSA key so NewGithubClient succeeds
	keyPEM := generateTestRSAKey(t)
	keyFile := filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0o600))

	prowCfg := "github:\n  app_id: 1\n  installation_id: 1\n  private_key_path: " + keyFile + "\nfeatures:\n  label_sync:\n    path: /nonexistent/labels.yml\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)

	LabelSync()
	assert.Equal(t, 1, *code)
}

func TestLabelSync_RunFails(t *testing.T) {
	code := captureExit(t)

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))

	keyPEM := generateTestRSAKey(t)
	keyFile := filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0o600))

	labelsPath := filepath.Join(dir, "labels.yml")
	require.NoError(t, os.WriteFile(labelsPath, []byte("categories: []\n"), 0o644))

	prowCfg := "github:\n  app_id: 1\n  installation_id: 1\n  private_key_path: " + keyFile + "\nfeatures:\n  label_sync:\n    path: " + labelsPath + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)

	// newGithubClientForLabelSync will create a real transport but API calls will fail
	// (no real GitHub). The Run will fail on listRepos.
	LabelSync()
	assert.Equal(t, 1, *code)
}
