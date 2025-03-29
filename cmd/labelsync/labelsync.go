package labelsync

import (
	"context"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/google/go-github/v69/github"
	"github.com/lmittmann/tint"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v3"
)

var logLevel = new(slog.LevelVar)

// https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28#primary-rate-limit-for-github-app-installations
// https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28#about-secondary-rate-limits
// It's difficult to determine the exact rate limit for a GitHub App installation, but it's generally around 5000 requests per hour.
// However, it does explictly state that the concurrent is 100. So we're doing 50 to be safe. There's almost certainly a better way to do this.
var maxConcurrentRequests = 50

type LabelSyncConfig struct {
	Overwrite  bool `yaml:"overwrite"`
	DryRun     bool `yaml:"dry_run"`
	Categories []struct {
		Name          string `yaml:"name"`
		CategoryColor string `yaml:"category_color,omitempty"`
		Labels        []struct {
			Name        string  `yaml:"name"`
			Color       string  `yaml:"color,omitempty"`
			Description *string `yaml:"description,omitempty"`
		} `yaml:"labels"`
	} `yaml:"categories"`
	ExtraLabels []struct {
		Name        string  `yaml:"name"`
		Description *string `yaml:"description,omitempty"`
		Color       string  `yaml:"color"`
	} `yaml:"extra_labels"`
}

func GetLabelSyncConfig(logger *slog.Logger) *LabelSyncConfig {
	lsc := &LabelSyncConfig{}

	yamlFile, err := os.ReadFile(".github/labels.yml")
	if err != nil {
		log.Fatalf("failed to read labels.yml: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, lsc)
	if err != nil {
		log.Fatalf("failed to unmarshal labels.yml: %v", err)
	}

	for _, category := range lsc.Categories {
		for _, label := range category.Labels {
			if category.CategoryColor == "" && label.Color == "" {
				log.Fatalf("either category_color or color must be set for category %s, label %s", category.Name, label.Name)
			}
		}
	}

	if lsc.Overwrite {
		logger.Warn("overwrite is enabled, existing labels with the same name will be overwritten")
	}
	if lsc.DryRun {
		logger.Info("dry run mode is enabled, no changes will be made")
	}

	logger.Debug("label sync config", slog.Any("config", lsc))
	logger.Debug("categories", slog.Any("categories", lsc.Categories))
	logger.Debug("extra_labels", slog.Any("extra_labels", lsc.ExtraLabels))

	return lsc
}

func syncLabels(owner string, labels []github.Label, overwrite bool, dryRun bool, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()

	repos, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{})
	if err != nil {
		logger.Error("failed to list repositories", slog.String("error", err.Error()))
		return
	}

	if len(repos.Repositories) == 0 {
		logger.Error("no repositories found", slog.String("owner", owner))
		return
	}

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(int64(maxConcurrentRequests))

	for _, repo := range repos.Repositories {
		if repo.GetArchived() {
			logger.Debug("skipping archived repo", slog.String("repo", *repo.FullName))
			continue
		}
		if repo.GetFork() {
			logger.Debug("skipping forked repo", slog.String("repo", *repo.FullName))
			continue
		}
		for _, label := range labels {
			wg.Add(1)
			if err := sem.Acquire(ctx, 1); err != nil {
				logger.Error("failed to acquire semaphore", slog.String("error", err.Error()))
				wg.Done()
				return
			}
			go func(repo *github.Repository, label github.Label) {
				defer wg.Done()
				defer sem.Release(1)

				labelFromRepo, _, err := client.Issues.GetLabel(ctx, owner, *repo.Name, *label.Name)
				if err != nil {
					logger.Debug("label not found, attempting create", slog.String("repo", *repo.FullName), slog.String("label", *label.Name))
					if dryRun {
						logger.Info("dry run: would create label", slog.String("repo", *repo.FullName), slog.String("label", *label.Name))
						return
					}
					_, _, err = client.Issues.CreateLabel(ctx, owner, *repo.Name, &label)
					if err != nil {
						logger.Error("failed to create label", slog.String("error", err.Error()))
					}
					return
				}

				if overwrite {
					needsUpdate := false
					if labelFromRepo.GetColor() != *label.Color {
						needsUpdate = true
					}
					if labelFromRepo.GetDescription() != "" && label.Description != nil && labelFromRepo.GetDescription() != *label.Description {
						needsUpdate = true
					}
					if labelFromRepo.GetDescription() == "" && label.Description != nil {
						needsUpdate = true
					}
					if needsUpdate {
						logger.Debug("updating label", slog.String("repo", *repo.FullName), slog.String("label", *label.Name))
						if dryRun {
							logger.Info("dry run: would update label", slog.String("repo", *repo.FullName), slog.String("label", *label.Name))
							return
						}
						_, _, err = client.Issues.EditLabel(ctx, owner, *repo.Name, *label.Name, &label)
						if err != nil {
							logger.Error("failed to update label", slog.String("error", err.Error()))
						}
					}
				} else {
					logger.Debug("label exists and overwrite is not enabled", slog.String("repo", *repo.FullName), slog.String("label", *label.Name))
				}
			}(repo, label)
		}
	}
	wg.Wait()
}

func createLabels(owner string, config *LabelSyncConfig, client *github.Client, logger *slog.Logger) {
	var labels []github.Label
	for _, category := range config.Categories {
		for _, label := range category.Labels {
			fullLabelName := category.Name + "/" + label.Name
			color := category.CategoryColor
			if label.Color != "" {
				color = label.Color
			}
			labels = append(labels, github.Label{
				Name:        github.Ptr(fullLabelName),
				Color:       github.Ptr(color),
				Description: label.Description,
			})
		}
	}
	syncLabels(owner, labels, config.Overwrite, config.DryRun, client, logger)
}

func createExtraLabels(owner string, config *LabelSyncConfig, client *github.Client, logger *slog.Logger) {
	var labels []github.Label
	for _, label := range config.ExtraLabels {
		labels = append(labels, github.Label{
			Name:        github.Ptr(label.Name),
			Color:       github.Ptr(label.Color),
			Description: label.Description,
		})
	}
	syncLabels(owner, labels, config.Overwrite, config.DryRun, client, logger)
}

func LabelSync() {
	logLevel.Set(slog.LevelDebug)
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: logLevel}))

	client, err := githubapi.NewGithubClient(logger)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(2)
	}

	logger.Info("starting label sync")
	lsc := GetLabelSyncConfig(logger)
	logger.Info("categories", slog.Any("categories", lsc.Categories))
	logger.Info("extra labels", slog.Any("extra_labels", lsc.ExtraLabels))
	createLabels("patrickblackjr", lsc, client.GetClient(logger), logger)
	createExtraLabels("patrickblackjr", lsc, client.GetClient(logger), logger)
}
