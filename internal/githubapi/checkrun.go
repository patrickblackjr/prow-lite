package githubapi

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/go-github/v69/github"
)

// CreateCheckRun creates or updates a check run
func CreateCheckRun(owner, repo, sha, conclusion string, name string, client *github.Client, logger *slog.Logger) (*github.CheckRun, error) {
	if sha == "" {
		logger.Warn("SHA is empty, skipping check run creation")
		return nil, errors.New("SHA is empty")
	}

	ctx := context.Background()
	checkRun, _, err := client.Checks.CreateCheckRun(ctx, owner, repo, github.CreateCheckRunOptions{
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
		return nil, err
	}

	logger.Info("Created check run", slog.Int64("check_run_id", checkRun.GetID()), slog.String("conclusion", conclusion))
	return checkRun, nil
}
