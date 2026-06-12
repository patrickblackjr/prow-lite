package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type ProwLiteConfig struct {
	GitHub struct {
		GitHubAppId          int64  `yaml:"app_id"`
		GitHubInstallationId int64  `yaml:"installation_id"`
		PrivateKeyPath       string `yaml:"private_key_path"`
	} `yaml:"github"`
	Features struct {
		LabelSync struct {
			Path string `yaml:"path"`
		} `yaml:"label_sync"`
		LGTM struct {
			MinApprovals *int `yaml:"min_approvals"`
		} `yaml:"lgtm"`
	} `yaml:"features"`
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
