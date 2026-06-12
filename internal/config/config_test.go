package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(content), 0o644))
	t.Chdir(dir)
}

func TestGetProwLiteConfig_Success(t *testing.T) {
	writeConfig(t, `
github:
  app_id: 42
  installation_id: 7
  private_key_path: /tmp/key.pem
features:
  label_sync:
    path: labels.yml
`)
	cfg, err := GetProwLiteConfig(slog.Default())
	require.NoError(t, err)
	assert.Equal(t, int64(42), cfg.GitHub.GitHubAppId)
	assert.Equal(t, int64(7), cfg.GitHub.GitHubInstallationId)
	assert.Equal(t, "/tmp/key.pem", cfg.GitHub.PrivateKeyPath)
	assert.Equal(t, "labels.yml", cfg.Features.LabelSync.Path)
}

func TestGetProwLiteConfig_FileNotFound(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := GetProwLiteConfig(slog.Default())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read prow-lite.yml")
}

func TestGetProwLiteConfig_InvalidYAML(t *testing.T) {
	writeConfig(t, `{invalid: [`)
	_, err := GetProwLiteConfig(slog.Default())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal prow-lite.yml")
}

func TestGetLabelCategories_Success(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	labelsPath := filepath.Join(dir, "labels.yml")
	require.NoError(t, os.WriteFile(labelsPath, []byte(`
categories:
  - name: kind
    labels:
      - name: bug
      - name: feature
  - name: priority
    labels:
      - name: critical
`), 0o644))
	prowCfg := "features:\n  label_sync:\n    path: " + labelsPath + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)

	cats, err := GetLabelCategories(slog.Default())
	require.NoError(t, err)
	assert.Equal(t, []string{"bug", "feature"}, cats["kind"])
	assert.Equal(t, []string{"critical"}, cats["priority"])
}

func TestGetLabelCategories_NoProwConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	cats, err := GetLabelCategories(slog.Default())
	assert.Nil(t, cats)
	assert.Error(t, err)
}

func TestGetLabelCategories_NoPath(t *testing.T) {
	writeConfig(t, "github:\n  app_id: 1\n")
	cats, err := GetLabelCategories(slog.Default())
	assert.Nil(t, cats)
	assert.NoError(t, err)
}

func TestGetLabelCategories_FileNotFound(t *testing.T) {
	writeConfig(t, "features:\n  label_sync:\n    path: /nonexistent/labels.yml\n")
	_, err := GetLabelCategories(slog.Default())
	assert.Error(t, err)
}

func TestGetLabelCategories_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	labelsPath := filepath.Join(dir, "labels.yml")
	require.NoError(t, os.WriteFile(labelsPath, []byte("{invalid: ["), 0o644))
	prowCfg := "features:\n  label_sync:\n    path: " + labelsPath + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)
	_, err := GetLabelCategories(slog.Default())
	assert.Error(t, err)
}
