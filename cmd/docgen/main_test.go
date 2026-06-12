package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patrickblackjr/prow-lite/cmd/labelsync"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender_WritesContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.md")
	render(path, "hello {{.}}", "world")
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))
}

func TestRender_PanicsOnBadPath(t *testing.T) {
	assert.Panics(t, func() {
		render("/nonexistent/dir/file.md", "{{.}}", nil)
	})
}

func TestRender_PanicsOnTemplateExecutionError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.md")
	assert.Panics(t, func() {
		render(path, `{{join . ","}}`, 42)
	})
}

func TestCommandPluginTemplate_LGTM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lgtm.md")
	render(path, commandPluginTmpl, githubapi.LGTMPlugin)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(got)

	assert.Contains(t, content, "AUTO-GENERATED")
	assert.Contains(t, content, "## Slash Commands")
	assert.Contains(t, content, "/lgtm")
	assert.Contains(t, content, "/approve")
	assert.Contains(t, content, "## Pull Request Lifecycle Events")
	assert.Contains(t, content, "`opened`")
	assert.Contains(t, content, "`reopened`")
	assert.Contains(t, content, "## Labels")
	assert.Contains(t, content, "`lgtm`")
	assert.Contains(t, content, "`do-not-merge`")
}

func TestCommandPluginTemplate_Assign(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assign.md")
	render(path, commandPluginTmpl, githubapi.AssignPlugin)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(got)

	assert.Contains(t, content, "AUTO-GENERATED")
	assert.Contains(t, content, "/assign")
	assert.Contains(t, content, "/unassign")
	assert.NotContains(t, content, "## Pull Request Lifecycle Events")
	assert.NotContains(t, content, "## Labels")
}

func TestCommandPluginTemplate_Label(t *testing.T) {
	path := filepath.Join(t.TempDir(), "label.md")
	render(path, commandPluginTmpl, githubapi.LabelPlugin)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(got)

	assert.Contains(t, content, "AUTO-GENERATED")
	assert.Contains(t, content, "/label")
	assert.Contains(t, content, "/unlabel")
	assert.Contains(t, content, "## Notes")
	assert.Contains(t, content, "labels.yml")
	assert.NotContains(t, content, "## Pull Request Lifecycle Events")
}

func TestLabelsyncTemplate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "labelsync.md")
	render(path, labelsyncTmpl, labelsync.Plugin)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(got)

	assert.Contains(t, content, "AUTO-GENERATED")
	assert.Contains(t, content, "Label sync plugin")
	assert.Contains(t, content, "## Configuration")
	assert.Contains(t, content, "`overwrite`")
	assert.Contains(t, content, "`prune`")
	assert.Contains(t, content, "`dry_run`")
	assert.Contains(t, content, "## Behavior")
	assert.Contains(t, content, "1.")
}

func TestMain_CreatesFiles(t *testing.T) {
	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })

	require.NoError(t, os.Chdir(t.TempDir()))
	main()

	for _, path := range []string{
		"docs/plugins/lgtm.md",
		"docs/plugins/assign.md",
		"docs/plugins/label.md",
		"docs/plugins/labelsync.md",
	} {
		content, err := os.ReadFile(path)
		require.NoError(t, err, "expected %s to be created", path)
		assert.True(t, strings.Contains(string(content), "AUTO-GENERATED"),
			"%s missing AUTO-GENERATED header", path)
	}
}
