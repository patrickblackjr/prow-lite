package githubapi

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRSAKey(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func writeProwConfig(t *testing.T, keyPath string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	cfg := "github:\n  app_id: 1\n  installation_id: 1\n  private_key_path: " + keyPath + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(cfg), 0o644))
	t.Chdir(dir)
}

func TestLoadPrivateKey_EnvVar(t *testing.T) {
	t.Setenv("PROW_GITHUB_PRIVATE_KEY", "key-content")
	got, err := loadPrivateKey("")
	require.NoError(t, err)
	assert.Equal(t, []byte("key-content"), got)
}

func TestLoadPrivateKey_File(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "key*.pem")
	require.NoError(t, err)
	_, err = f.WriteString("file-key")
	require.NoError(t, err)
	_ = f.Close()

	got, err := loadPrivateKey(f.Name())
	require.NoError(t, err)
	assert.Equal(t, []byte("file-key"), got)
}

func TestLoadPrivateKey_NoConfig(t *testing.T) {
	t.Setenv("PROW_GITHUB_PRIVATE_KEY", "")
	_, err := loadPrivateKey("")
	assert.Error(t, err)
}

func TestNewGithubClient_Success(t *testing.T) {
	keyPEM := generateRSAKey(t)
	keyFile := filepath.Join(t.TempDir(), "key.pem")
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0o600))
	writeProwConfig(t, keyFile)

	client, err := NewGithubClient(slog.Default())
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewGithubClient_NoConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := NewGithubClient(slog.Default())
	assert.Error(t, err)
}

func TestNewGithubClient_NoPrivateKey(t *testing.T) {
	// Config exists but no private_key_path and no env var → loadPrivateKey errors.
	t.Setenv("PROW_GITHUB_PRIVATE_KEY", "")
	writeProwConfig(t, "") // empty private_key_path
	_, err := NewGithubClient(slog.Default())
	assert.Error(t, err)
}

func TestNewGithubClient_BadKey(t *testing.T) {
	keyFile := filepath.Join(t.TempDir(), "key.pem")
	require.NoError(t, os.WriteFile(keyFile, []byte("not-a-valid-key"), 0o600))
	writeProwConfig(t, keyFile)

	_, err := NewGithubClient(slog.Default())
	assert.Error(t, err)
}

func TestProwLiteGitHubClient_GetClient(t *testing.T) {
	inner := github.NewClient(nil)
	c := &ProwLiteGitHubClient{client: inner}
	assert.Equal(t, inner, c.GetClient())
}

func TestProwLiteGitHubClient_CreateCheckRun(t *testing.T) {
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))},
		),
	))
	c := &ProwLiteGitHubClient{client: mockClient}
	cr, _, err := c.CreateCheckRun(context.Background(), "owner", "repo", github.CreateCheckRunOptions{
		Name:    "test",
		HeadSHA: "abc123",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), cr.GetID())
}
