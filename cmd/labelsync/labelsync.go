package labelsync

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/google/go-github/v71/github"
	"github.com/lmittmann/tint"
	"github.com/patrickblackjr/prow-lite/internal/config"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v3"
)

// https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28#about-secondary-rate-limits
var maxConcurrentRequests int64 = 50

var osExit = os.Exit

type githubClientProvider interface {
	GetClient() *github.Client
}

var newGithubClientFunc func(*slog.Logger) (githubClientProvider, error) = func(logger *slog.Logger) (githubClientProvider, error) {
	return githubapi.NewGithubClient(logger)
}

type Category struct {
	Name          string          `yaml:"name"`
	CategoryColor string          `yaml:"category_color,omitempty"`
	Labels        []CategoryLabel `yaml:"labels"`
}

type CategoryLabel struct {
	Name        string  `yaml:"name"`
	Color       string  `yaml:"color,omitempty"`
	Description *string `yaml:"description,omitempty"`
}

type ExtraLabel struct {
	Name        string  `yaml:"name"`
	Description *string `yaml:"description,omitempty"`
	Color       string  `yaml:"color"`
}

type LabelSyncConfig struct {
	Overwrite   bool        `yaml:"overwrite"`
	Prune       bool        `yaml:"prune"` // Prune deletes labels not in the labels.yml file
	DryRun      bool        `yaml:"dry_run"`
	Categories  []Category  `yaml:"categories"`
	ExtraLabels []ExtraLabel `yaml:"extra_labels"`
}

func GetLabelSyncConfig(logger *slog.Logger) (*LabelSyncConfig, error) {
	cfg, err := config.GetProwLiteConfig(logger)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cfg.Features.LabelSync.Path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", cfg.Features.LabelSync.Path, err)
	}

	lsc := &LabelSyncConfig{}
	if err := yaml.Unmarshal(data, lsc); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", cfg.Features.LabelSync.Path, err)
	}

	for _, cat := range lsc.Categories {
		for _, label := range cat.Labels {
			if cat.CategoryColor == "" && label.Color == "" {
				return nil, fmt.Errorf("either category_color or color must be set for category %s, label %s", cat.Name, label.Name)
			}
		}
	}

	if lsc.Overwrite {
		logger.Warn("overwrite is enabled, existing labels with the same name will be overwritten")
	}
	if lsc.DryRun {
		logger.Info("dry run mode is enabled, no changes will be made")
	}

	logger.Debug("label sync config loaded",
		slog.Any("categories", lsc.Categories),
		slog.Any("extra_labels", lsc.ExtraLabels),
	)

	return lsc, nil
}

// Syncer executes a label sync run against all accessible repositories.
type Syncer struct {
	client *github.Client
	cfg    *LabelSyncConfig
	logger *slog.Logger
}

func NewSyncer(cfg *LabelSyncConfig, client *github.Client, logger *slog.Logger) *Syncer {
	return &Syncer{cfg: cfg, client: client, logger: logger}
}

func (s *Syncer) Run(ctx context.Context) error {
	repos, err := s.listRepos(ctx)
	if err != nil {
		return err
	}

	labels := s.buildLabels()
	if err := s.syncLabels(ctx, repos, labels); err != nil {
		return err
	}

	if s.cfg.Prune {
		return s.pruneLabels(ctx, repos, labels)
	}
	return nil
}

func (s *Syncer) listRepos(ctx context.Context) ([]*github.Repository, error) {
	result, _, err := s.client.Apps.ListRepos(ctx, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	if len(result.Repositories) == 0 {
		return nil, fmt.Errorf("no repositories found for installation")
	}

	repos := make([]*github.Repository, 0, len(result.Repositories))
	for _, r := range result.Repositories {
		if r.GetArchived() {
			s.logger.Debug("skipping archived repo", slog.String("repo", r.GetFullName()))
			continue
		}
		if r.GetFork() {
			s.logger.Debug("skipping forked repo", slog.String("repo", r.GetFullName()))
			continue
		}
		repos = append(repos, r)
	}
	return repos, nil
}

func (s *Syncer) buildLabels() []github.Label {
	var labels []github.Label

	for _, cat := range s.cfg.Categories {
		for _, l := range cat.Labels {
			color := cat.CategoryColor
			if l.Color != "" {
				color = l.Color
			}
			labels = append(labels, github.Label{
				Name:        github.Ptr(cat.Name + "/" + l.Name),
				Color:       github.Ptr(color),
				Description: l.Description,
			})
		}
	}

	for _, l := range s.cfg.ExtraLabels {
		labels = append(labels, github.Label{
			Name:        github.Ptr(l.Name),
			Color:       github.Ptr(l.Color),
			Description: l.Description,
		})
	}

	return labels
}

func (s *Syncer) syncLabels(ctx context.Context, repos []*github.Repository, labels []github.Label) error {
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(maxConcurrentRequests)

	for _, repo := range repos {
		owner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()
		fullName := repo.GetFullName()

		for _, label := range labels {
			wg.Add(1)
			if err := sem.Acquire(ctx, 1); err != nil {
				wg.Done()
				return fmt.Errorf("acquire semaphore: %w", err)
			}
			go func(owner, repoName, fullName string, label github.Label) {
				defer wg.Done()
				defer sem.Release(1)
				s.syncLabel(ctx, owner, repoName, fullName, label)
			}(owner, repoName, fullName, label)
		}
	}

	wg.Wait()
	return nil
}

func (s *Syncer) syncLabel(ctx context.Context, owner, repoName, fullName string, label github.Label) {
	existing, _, err := s.client.Issues.GetLabel(ctx, owner, repoName, label.GetName())
	if err != nil {
		s.logger.Debug("label not found, creating", slog.String("repo", fullName), slog.String("label", label.GetName()))
		if s.cfg.DryRun {
			s.logger.Info("dry run: would create label", slog.String("repo", fullName), slog.String("label", label.GetName()))
			return
		}
		if _, _, err := s.client.Issues.CreateLabel(ctx, owner, repoName, &label); err != nil {
			s.logger.Error("failed to create label", slog.String("repo", fullName), slog.String("label", label.GetName()), slog.String("error", err.Error()))
		}
		return
	}

	if !s.cfg.Overwrite {
		s.logger.Debug("label exists, overwrite disabled", slog.String("repo", fullName), slog.String("label", label.GetName()))
		return
	}

	if !labelNeedsUpdate(existing, &label) {
		return
	}

	s.logger.Debug("updating label", slog.String("repo", fullName), slog.String("label", label.GetName()))
	if s.cfg.DryRun {
		s.logger.Info("dry run: would update label", slog.String("repo", fullName), slog.String("label", label.GetName()))
		return
	}
	if _, _, err := s.client.Issues.EditLabel(ctx, owner, repoName, label.GetName(), &label); err != nil {
		s.logger.Error("failed to update label", slog.String("repo", fullName), slog.String("label", label.GetName()), slog.String("error", err.Error()))
	}
}

func labelNeedsUpdate(existing, desired *github.Label) bool {
	if existing.GetColor() != desired.GetColor() {
		return true
	}
	// Only update description when the desired config specifies one and it differs.
	if desired.Description != nil && existing.GetDescription() != desired.GetDescription() {
		return true
	}
	return false
}

func (s *Syncer) pruneLabels(ctx context.Context, repos []*github.Repository, keep []github.Label) error {
	keepSet := make(map[string]struct{}, len(keep))
	for _, l := range keep {
		keepSet[l.GetName()] = struct{}{}
	}

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(maxConcurrentRequests)

	for _, repo := range repos {
		owner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()
		fullName := repo.GetFullName()

		wg.Add(1)
		if err := sem.Acquire(ctx, 1); err != nil {
			wg.Done()
			return fmt.Errorf("acquire semaphore: %w", err)
		}
		go func(owner, repoName, fullName string) {
			defer wg.Done()
			defer sem.Release(1)
			s.pruneRepo(ctx, owner, repoName, fullName, keepSet)
		}(owner, repoName, fullName)
	}

	wg.Wait()
	return nil
}

func (s *Syncer) pruneRepo(ctx context.Context, owner, repoName, fullName string, keepSet map[string]struct{}) {
	var allLabels []*github.Label
	opts := &github.ListOptions{PerPage: 100}
	for {
		page, resp, err := s.client.Issues.ListLabels(ctx, owner, repoName, opts)
		if err != nil {
			s.logger.Error("failed to list labels", slog.String("repo", fullName), slog.String("error", err.Error()))
			return
		}
		allLabels = append(allLabels, page...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	for _, label := range allLabels {
		if _, ok := keepSet[label.GetName()]; ok {
			continue
		}
		s.logger.Debug("pruning label", slog.String("repo", fullName), slog.String("label", label.GetName()))
		if s.cfg.DryRun {
			s.logger.Info("dry run: would delete label", slog.String("repo", fullName), slog.String("label", label.GetName()))
			continue
		}
		if _, err := s.client.Issues.DeleteLabel(ctx, owner, repoName, label.GetName()); err != nil {
			s.logger.Error("failed to delete label", slog.String("repo", fullName), slog.String("label", label.GetName()), slog.String("error", err.Error()))
		} else {
			s.logger.Info("deleted label", slog.String("repo", fullName), slog.String("label", label.GetName()))
		}
	}

	s.logger.Debug("finished pruning repo", slog.String("repo", fullName))
}

func LabelSync() {
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelDebug)
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: logLevel}))

	ctx := context.Background()

	ghClient, err := newGithubClientFunc(logger)
	if err != nil {
		logger.Error("failed to create GitHub client", slog.String("error", err.Error()))
		osExit(2)
		return
	}

	lsc, err := GetLabelSyncConfig(logger)
	if err != nil {
		logger.Error("failed to load label sync config", slog.String("error", err.Error()))
		osExit(1)
		return
	}

	logger.Info("starting label sync",
		slog.Int("categories", len(lsc.Categories)),
		slog.Int("extra_labels", len(lsc.ExtraLabels)),
	)

	syncer := NewSyncer(lsc, ghClient.GetClient(), logger)
	if err := syncer.Run(ctx); err != nil {
		logger.Error("label sync failed", slog.String("error", err.Error()))
		osExit(1)
		return
	}

	logger.Info("label sync complete")
}