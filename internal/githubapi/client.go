package githubapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v71/github"
	"github.com/patrickblackjr/prow-lite/internal/config"
)

// ProwLiteGitHubClient wraps a GitHub client authenticated as a GitHub App installation.
type ProwLiteGitHubClient struct {
	client *github.Client
}

// ProwGitHubClient is the interface consumed by callers that need a GitHub client.
type ProwGitHubClient interface {
	GetClient() *github.Client
	CreateCheckRun(ctx context.Context, owner, repo string, opt github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error)
}

func NewGithubClient(logger *slog.Logger) (*ProwLiteGitHubClient, error) {
	cfg, err := config.GetProwLiteConfig(logger)
	if err != nil {
		return nil, err
	}

	privateKey, err := loadPrivateKey(cfg.GitHub.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	itr, err := ghinstallation.New(http.DefaultTransport, cfg.GitHub.GitHubAppId, cfg.GitHub.GitHubInstallationId, privateKey)
	if err != nil {
		logger.Error("failed to create GitHub App client")
		return nil, err
	}
	return &ProwLiteGitHubClient{client: github.NewClient(&http.Client{Transport: itr})}, nil
}

// loadPrivateKey reads the GitHub App private key. It checks, PROW_GITHUB_PRIVATE_KEY env var
// followed by the path from prow-lite.yml github.private_key_path
func loadPrivateKey(cfgPath string) ([]byte, error) {
	if raw := os.Getenv("PROW_GITHUB_PRIVATE_KEY"); raw != "" {
		return []byte(raw), nil
	}
	if cfgPath != "" {
		return os.ReadFile(cfgPath)
	}
	return nil, fmt.Errorf("no private key configured: set github.private_key_path in prow-lite.yml or PROW_GITHUB_PRIVATE_KEY env var")
}

func (g *ProwLiteGitHubClient) GetClient() *github.Client {
	return g.client
}

func (g *ProwLiteGitHubClient) CreateCheckRun(ctx context.Context, owner, repo string, opt github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	return g.client.Checks.CreateCheckRun(ctx, owner, repo, opt)
}
