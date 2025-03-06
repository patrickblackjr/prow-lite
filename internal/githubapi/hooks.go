package githubapi

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v68/github"
)

func EnsureLabels(owner string, labels []string, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()

	repos, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{})
	if err != nil {
		logger.Debug("no repos found", slog.String("owner", owner))
	}

	for _, repo := range repos.Repositories {
		label, _, err := client.Issues.GetLabel(ctx, owner, *repo.Name, "lgtm")
		logger.Debug("found label", slog.String("repo", *repo.FullName), slog.String("label", *label.Name))
		if err != nil {
			_, _, err = client.Issues.CreateLabel(ctx, owner, *repo.Name, &github.Label{
				Name:        github.Ptr("lgtm"),
				Color:       github.Ptr("0e8a16"),
				Description: github.Ptr("Approved by reviewers"),
			})
			if err != nil {
				logger.Error("failed to create lgtm label", slog.String("error", err.Error()))
				return
			}
			logger.Info("created lgtm label", slog.String("repo", *repo.FullName))
		}
	}

}
