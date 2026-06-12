package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type ProwLiteConfig struct {
	// GitHub contains the configuration needed to authenticate with the GitHub
	// API as a GitHub App.
	GitHub struct {
		// GitHubAppId is the ID of the GitHub App as registered in GitHub.
		GitHubAppId int64 `yaml:"app_id"`
		// GitHubInstallationId is the ID of the specific installation of the
		// GitHub App. This is required to obtain an access token scoped to the
		// repositories where the app is installed.
		GitHubInstallationId int64 `yaml:"installation_id"`
		// PrivateKeyPath is the filesystem path to the PEM-encoded private key
		// for the GitHub App. This is used to authenticate as the app and obtain
		// access tokens for API calls.
		PrivateKeyPath string `yaml:"private_key_path"`
	} `yaml:"github"`
	// Features contains configuration for features of Prow Lite.
	Features struct {
		// LabelSync contains configuration for the label sync feature, which
		// keeps repository labels in sync with a canonical set defined in the
		// repository.
		LabelSync struct {
			// Path is the filesystem path to the YAML file that defines the
			// canonical set of labels for the repository.
			Path string `yaml:"path"`
		} `yaml:"label_sync"`
		// LGTM contains configuration for the LGTM approval feature, which
		// allows users to approve PRs with slash commands and tracks the number
		// of approvals in a check run.
		LGTM struct {
			// MinApprovals is the number of approvals required for a PR to be
			// considered approved.
			MinApprovals *int `yaml:"min_approvals"`
		} `yaml:"lgtm"`
	} `yaml:"features"`
}

// GetLabelCategories parses the labels.yml file configured under features.label_sync.path
// and returns a map of category name → slice of label names. Returns nil without error
// if the label_sync path is not configured or the prow-lite config cannot be loaded.
func GetLabelCategories(logger *slog.Logger) (map[string][]string, error) {
	plc, err := GetProwLiteConfig(logger)
	if err != nil || plc.Features.LabelSync.Path == "" {
		return nil, err
	}

	data, err := os.ReadFile(plc.Features.LabelSync.Path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", plc.Features.LabelSync.Path, err)
	}

	var cfg struct {
		Categories []struct {
			Name   string `yaml:"name"`
			Labels []struct {
				Name string `yaml:"name"`
			} `yaml:"labels"`
		} `yaml:"categories"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", plc.Features.LabelSync.Path, err)
	}

	result := make(map[string][]string, len(cfg.Categories))
	for _, cat := range cfg.Categories {
		names := make([]string, len(cat.Labels))
		for i, l := range cat.Labels {
			names[i] = l.Name
		}
		result[cat.Name] = names
	}
	return result, nil
}

func GetProwLiteConfig(_ *slog.Logger) (*ProwLiteConfig, error) {
	plc := &ProwLiteConfig{}

	yamlFile, err := os.ReadFile(".github/prow-lite.yml")
	if err != nil {
		return nil, fmt.Errorf("read prow-lite.yml: %w", err)
	}

	if err = yaml.Unmarshal(yamlFile, plc); err != nil {
		return nil, fmt.Errorf("unmarshal prow-lite.yml: %w", err)
	}

	return plc, nil
}
