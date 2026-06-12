package event

import (
	"log/slog"

	"github.com/google/go-github/v71/github"
	"github.com/patrickblackjr/prow-lite/internal/config"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
)

// NewProcessComment loads min_approvals and label categories from the repository's config
// files and returns a configured ProcessComment function. Config errors are non-fatal;
// defaults are used when config is missing or invalid.
func NewProcessComment(logger *slog.Logger) func(*github.IssueCommentEvent, *github.Client, *slog.Logger) {
	categories, _ := config.GetLabelCategories(logger)
	return githubapi.NewProcessComment(minApprovalsFromConfig(logger), categories)
}

// NewPREventHandler loads min_approvals from the repository's config file and returns
// a configured pull_request event handler. Defaults to 1 approval if config is missing or invalid.
func NewPREventHandler(logger *slog.Logger) func(*github.PullRequestEvent, *github.Client, *slog.Logger) {
	return githubapi.NewPREventHandler(minApprovalsFromConfig(logger))
}

func minApprovalsFromConfig(logger *slog.Logger) int {
	minApprovals := 1
	if cfg, err := config.GetProwLiteConfig(logger); err == nil && cfg.Features.LGTM.MinApprovals != nil {
		minApprovals = *cfg.Features.LGTM.MinApprovals
	}
	return minApprovals
}
