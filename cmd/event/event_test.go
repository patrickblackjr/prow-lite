package event

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func writeProwConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(content), 0o644))
	t.Chdir(dir)
}

func TestNewProcessComment_NoConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.NotNil(t, NewProcessComment(discardLogger()))
}

func TestNewProcessComment_WithMinApprovals(t *testing.T) {
	writeProwConfig(t, "features:\n  lgtm:\n    min_approvals: 2\n")
	assert.NotNil(t, NewProcessComment(discardLogger()))
}

func TestNewProcessComment_WithLabelCategories(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github"), 0o755))
	labelsPath := filepath.Join(dir, "labels.yml")
	require.NoError(t, os.WriteFile(labelsPath, []byte("categories:\n  - name: kind\n    labels:\n      - name: bug\n"), 0o644))
	prowCfg := "features:\n  lgtm:\n    min_approvals: 1\n  label_sync:\n    path: " + labelsPath + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github", "prow-lite.yml"), []byte(prowCfg), 0o644))
	t.Chdir(dir)
	assert.NotNil(t, NewProcessComment(discardLogger()))
}

func TestNewPREventHandler_NoConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.NotNil(t, NewPREventHandler(discardLogger()))
}

func TestNewPREventHandler_WithConfig(t *testing.T) {
	writeProwConfig(t, "features:\n  lgtm:\n    min_approvals: 2\n")
	assert.NotNil(t, NewPREventHandler(discardLogger()))
}
