package labelsync

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/google/go-github/v69/github"
	"github.com/lmittmann/tint"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"gopkg.in/yaml.v3"
)

var logLevel = new(slog.LevelVar)

type LabelSyncConfig struct {
	Overwrite bool `yaml:"overwrite"`
	DryRun    bool `yaml:"dry_run"`
	Areas     struct {
		CategoryColor string `yaml:"category_color"`
		Labels        []struct {
			Name        string  `yaml:"name"`
			Description *string `yaml:"description,omitempty"`
		} `yaml:"labels"`
	} `yaml:"areas"`
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

	if lsc.Overwrite {
		logger.Warn("overwrite is enabled, existing labels with the same name will be overwritten")
	}
	if lsc.DryRun {
		logger.Info("dry run mode is enabled, no changes will be made")
	}

	logger.Debug("label sync config", slog.Any("config", lsc))
	logger.Debug("areas", slog.Any("areas", lsc.Areas))
	logger.Debug("extra_labels", slog.Any("extra_labels", lsc.ExtraLabels))

	return lsc
}

func createLabels(owner string, labels *LabelSyncConfig, client *github.Client, logger *slog.Logger) {
	ctx := context.Background()

	repos, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{})
	if err != nil {
		logger.Debug("no repos found", slog.String("owner", owner))
		return
	}

	for _, repo := range repos.Repositories {
		for _, label := range labels.ExtraLabels {
			labelFromRepo, _, err := client.Issues.GetLabel(ctx, owner, *repo.Name, label.Name)
			if err != nil {
				logger.Debug("label not found, attempting create", slog.String("repo", *repo.FullName), slog.String("label", label.Name))
				if labels.DryRun {
					logger.Info("dry run: would create label", slog.String("repo", *repo.FullName), slog.String("label", label.Name))
					continue
				}
				newLabelStruct := &github.Label{
					Name:        github.Ptr(label.Name),
					Color:       github.Ptr(label.Color),
					Description: label.Description,
				}
				_, _, err = client.Issues.CreateLabel(ctx, owner, *repo.Name, newLabelStruct)
				if err != nil {
					logger.Error("failed to create label", slog.String("error", err.Error()))
				}
				continue
			}

			if labels.Overwrite {
				needsUpdate := false
				if labelFromRepo.GetColor() != label.Color {
					needsUpdate = true
				}
				if labelFromRepo.GetDescription() != "" && label.Description != nil && labelFromRepo.GetDescription() != *label.Description {
					needsUpdate = true
				}
				if labelFromRepo.GetDescription() == "" && label.Description != nil {
					needsUpdate = true
				}
				if needsUpdate {
					logger.Debug("updating label", slog.String("repo", *repo.FullName), slog.String("label", label.Name))
					if labels.DryRun {
						logger.Info("dry run: would update label", slog.String("repo", *repo.FullName), slog.String("label", label.Name))
						continue
					}
					updatedLabelStruct := &github.Label{
						Name:        github.Ptr(label.Name),
						Color:       github.Ptr(label.Color),
						Description: label.Description,
					}
					_, _, err = client.Issues.EditLabel(ctx, owner, *repo.Name, label.Name, updatedLabelStruct)
					if err != nil {
						logger.Error("failed to update label", slog.String("error", err.Error()))
					}
				}
			} else {
				logger.Debug("label exists and overwrite is not enabled", slog.String("repo", *repo.FullName), slog.String("label", label.Name))
			}
		}
	}
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
	logger.Info("labels", slog.Any("labels", lsc.ExtraLabels))
	createLabels("patrickblackjr", lsc, client.GetClient(logger), logger)
}
