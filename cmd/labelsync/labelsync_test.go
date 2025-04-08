package labelsync

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v70/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGitHubClient is a mock implementation of the GitHub client.
type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) AppsListRepos(ctx context.Context, opts *github.ListOptions) (*github.ListRepositories, *github.Response, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*github.ListRepositories), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) IssuesGetLabel(ctx context.Context, owner, repo, name string) (*github.Label, *github.Response, error) {
	args := m.Called(ctx, owner, repo, name)
	return args.Get(0).(*github.Label), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) IssuesCreateLabel(ctx context.Context, owner, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	args := m.Called(ctx, owner, repo, label)
	return args.Get(0).(*github.Label), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) IssuesEditLabel(ctx context.Context, owner, repo, name string, label *github.Label) (*github.Label, *github.Response, error) {
	args := m.Called(ctx, owner, repo, name, label)
	return args.Get(0).(*github.Label), args.Get(1).(*github.Response), args.Error(2)
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestGetLabelSyncConfig_Success(t *testing.T) {
	logger := setupLogger()
	configYAML := `
overwrite: true
dry_run: true
categories:
  - name: "bug"
    category_color: "ff0000"
    labels:
      - name: "critical"
        color: "ff0000"
        description: "Critical bugs"
extra_labels:
  - name: "enhancement"
    color: "00ff00"
    description: "Feature enhancements"
`
	_ = os.WriteFile(".github/labels.yml", []byte(configYAML), 0644)
	defer os.Remove(".github/labels.yml")

	config := GetLabelSyncConfig(logger)

	assert.True(t, config.Overwrite)
	assert.True(t, config.DryRun)
	assert.Equal(t, 1, len(config.Categories))
	assert.Equal(t, "bug", config.Categories[0].Name)
	assert.Equal(t, "ff0000", config.Categories[0].CategoryColor)
	assert.Equal(t, 1, len(config.Categories[0].Labels))
	assert.Equal(t, "critical", config.Categories[0].Labels[0].Name)
	assert.Equal(t, "ff0000", config.Categories[0].Labels[0].Color)
	assert.Equal(t, "Critical bugs", *config.Categories[0].Labels[0].Description)
	assert.Equal(t, 1, len(config.ExtraLabels))
	assert.Equal(t, "enhancement", config.ExtraLabels[0].Name)
	assert.Equal(t, "00ff00", config.ExtraLabels[0].Color)
	assert.Equal(t, "Feature enhancements", *config.ExtraLabels[0].Description)
}

func TestSyncLabels_CreateLabel(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "not found", http.StatusNotFound)
			}),
		),
		mock.WithRequestMatch(
			mock.PostReposIssuesByOwnerByRepo,
			github.Label{Name: github.Ptr("bug/critical"), Color: github.Ptr("ff0000")},
		),
	))

	label := github.Label{Name: github.Ptr("bug/critical"), Color: github.Ptr("ff0000")}
	syncLabels("owner", []github.Label{label}, false, false, mockClient, logger)
	// No assertions needed as we are testing for no errors in execution
}

func TestSyncLabels_UpdateLabel(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposIssuesLabelsByOwnerByRepoByIssueNumber,
			github.Label{Name: github.Ptr("bug/critical"), Color: github.Ptr("000000")},
		),
		mock.WithRequestMatch(
			mock.PatchReposIssuesByOwnerByRepoByIssueNumber,
			github.Label{Name: github.Ptr("bug/critical"), Color: github.Ptr("ff0000")},
		),
	))

	label := github.Label{Name: github.Ptr("bug/critical"), Color: github.Ptr("ff0000")}
	syncLabels("owner", []github.Label{label}, true, false, mockClient, logger)
	// No assertions needed as we are testing for no errors in execution
}

func TestCreateLabels_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposIssuesLabelsByOwnerByRepo,
			github.Label{Name: github.Ptr("bug/critical"), Color: github.Ptr("ff0000")},
		),
	))

	config := &LabelSyncConfig{
		Categories: []struct {
			Name          string `yaml:"name"`
			CategoryColor string `yaml:"category_color,omitempty"`
			Labels        []struct {
				Name        string  `yaml:"name"`
				Color       string  `yaml:"color,omitempty"`
				Description *string `yaml:"description,omitempty"`
			} `yaml:"labels"`
		}{
			{
				Name:          "bug",
				CategoryColor: "ff0000",
				Labels: []struct {
					Name        string  `yaml:"name"`
					Color       string  `yaml:"color,omitempty"`
					Description *string `yaml:"description,omitempty"`
				}{
					{Name: "critical", Color: "ff0000", Description: github.Ptr("Critical bugs")},
				},
			},
		},
	}

	createLabels("owner", config, mockClient, logger)
	// No assertions needed as we are testing for no errors in execution
}

func TestCreateExtraLabels_Success(t *testing.T) {
	logger := setupLogger()
	mockClient := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposIssuesLabelsByOwnerByRepo,
			github.Label{Name: github.Ptr("enhancement"), Color: github.Ptr("00ff00")},
		),
	))

	config := &LabelSyncConfig{
		ExtraLabels: []struct {
			Name        string  `yaml:"name"`
			Description *string `yaml:"description,omitempty"`
			Color       string  `yaml:"color"`
		}{
			{Name: "enhancement", Color: "00ff00", Description: github.Ptr("Feature enhancements")},
		},
	}

	createExtraLabels("owner", config, mockClient, logger)
	// No assertions needed as we are testing for no errors in execution
}
