package checkrun

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v69/github"
)

// ProwGitHubClient is an interface for the GitHub client
type ProwGitHubClient interface {
	CreateCheckRun(ctx context.Context, owner, repo string, opt github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error)
}

// CreateCheckRun creates or updates a check run
func CreateCheckRun(owner, repo, sha, conclusion string, name string, client ProwGitHubClient, logger *slog.Logger) {
	if sha == "" {
		logger.Warn("SHA is empty, skipping check run creation")
		return
	}

	ctx := context.Background()
	checkRun, _, err := client.CreateCheckRun(ctx, owner, repo, github.CreateCheckRunOptions{
		Name:       "LGTM",
		HeadSHA:    sha,
		Status:     github.Ptr("completed"),
		Conclusion: github.Ptr(conclusion),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(name),
			Summary: github.Ptr("This PR is not approved."),
		},
	})
	if err != nil {
		logger.Error("Failed to create check run", slog.String("error", err.Error()))
		return
	}
	logger.Info("Created check run", slog.Int64("check_run_id", checkRun.GetID()), slog.String("conclusion", conclusion))
}
