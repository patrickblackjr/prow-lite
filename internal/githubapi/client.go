package githubapi

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v71/github"
	"github.com/patrickblackjr/prow-lite/internal/config"
)

// ProwLiteGitHubClient is a struct for the GitHub client
type ProwLiteGitHubClient struct {
	client *github.Client
}

// ProwGitHubClient is an interface for the GitHub client
type ProwGitHubClient interface {
	GetClient(logger *slog.Logger) *github.Client
	CreateCheckRun(ctx context.Context, owner, repo string, opt github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error)
}

func NewGithubClient(logger *slog.Logger) (*ProwLiteGitHubClient, error) {
	config := config.GetProwLiteConfig(logger)
	privateKey := os.Getenv("PROW_GITHUB_PRIVATE_KEY")

	itr, err := ghinstallation.New(http.DefaultTransport, config.GitHub.GitHubAppId, config.GitHub.GitHubInstallationId, []byte(privateKey))
	if err != nil {
		logger.Error("failed to create github app client")
		return nil, err
	}
	client := github.NewClient(&http.Client{Transport: itr})
	return &ProwLiteGitHubClient{client: client}, nil
}

func (g *ProwLiteGitHubClient) GetClient(logger *slog.Logger) *github.Client {
	return g.client
}

func (g *ProwLiteGitHubClient) CreateCheckRun(ctx context.Context, owner, repo string, opt github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	return g.client.Checks.CreateCheckRun(ctx, owner, repo, opt)
}

func (g *ProwLiteGitHubClient) GetPRSHA(owner, repo string, prNumber int, logger *slog.Logger) string {
	ctx := context.Background()
	pr, _, err := g.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		logger.Error("Failed to retrieve PR SHA", slog.String("error", err.Error()))
		return ""
	}
	return pr.GetHead().GetSHA()
}

func (g *ProwLiteGitHubClient) RemoveLabel(owner, repo string, prNumber int, label string, logger *slog.Logger) {
	ctx := context.Background()
	_, err := g.client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNumber, label)
	if err != nil {
		logger.Warn("Failed to remove label", slog.String("label", label), slog.String("error", err.Error()))
		return
	}
	logger.Info("Removed label", slog.String("label", label))
}

func (g *ProwLiteGitHubClient) AddComment(owner, repo string, prNumber int, commentText string, logger *slog.Logger) {
	ctx := context.Background()
	_, _, err := g.client.Issues.CreateComment(ctx, owner, repo, prNumber, &github.IssueComment{Body: github.Ptr(commentText)})
	if err != nil {
		logger.Error("failed to add comment", slog.String("comment_text", commentText))
		return
	}
	logger.Info("added comment to PR", slog.String("comment_text", commentText))
}

func (g *ProwLiteGitHubClient) CreateLabel(owner, repo, name, color, description string, logger *slog.Logger) {
	ctx := context.Background()
	_, _, err := g.client.Issues.CreateLabel(ctx, owner, repo, &github.Label{
		Name:        github.Ptr(name),
		Color:       github.Ptr(color),
		Description: github.Ptr(description),
	})
	if err != nil {
		logger.Error("failed to create label", slog.String("error", err.Error()))
		return
	}
	logger.Info("created label", slog.String("name", name))
}

func (g *ProwLiteGitHubClient) AddLabelsToIssue(owner, repo string, prNumber int, labels []string, logger *slog.Logger) {
	ctx := context.Background()
	_, _, err := g.client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, labels)
	if err != nil {
		logger.Error("failed to add labels", slog.String("error", err.Error()))
		return
	}
	logger.Info("added labels", slog.String("labels", strings.Join(labels, ",")))
}
