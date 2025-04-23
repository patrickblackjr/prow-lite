package config

import (
	"log"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type ProwLiteConfig struct {
	GitHub struct {
		GitHubAppId          int64 `yaml:"app_id"`
		GitHubInstallationId int64 `yaml:"installation_id"`
	} `yaml:"github"`
	Features struct {
		LabelSync struct {
			Path string `yaml:"path"`
		} `yaml:"label_sync"`
	} `yaml:"features"`
}

func GetProwLiteConfig(logger *slog.Logger) *ProwLiteConfig {
	plc := &ProwLiteConfig{}

	yamlFile, err := os.ReadFile(".github/prow-lite.yml")
	if err != nil {
		log.Fatalf("failed to read prow-lite.yml: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, plc)
	if err != nil {
		log.Fatalf("failed to unmarshal prow-lite.yml: %v", err)
	}

	return plc
}
