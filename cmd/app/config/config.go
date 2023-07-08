package config

import (
	"github.com/google/go-github/v50/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config is global object that holds all application level variables
var Config appConfig

type appConfig struct {
	ServerPort           int `mapstructure:"server_port"`
	GitHubClient         *github.Client
	GitHubWebhookSecret  string `mapstructure:"github_webhook_secret"`
	GitHubInstallationID int64  `mapstructure:"installation_id"`
}

// LoadConfig loads config from files
func LoadConfig(configPaths ...string) error {
	v := viper.New()
	v.SetConfigName("server")
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PROW")
	v.AutomaticEnv()

	v.SetDefault("server_port", 8080)

	for _, path := range configPaths {
		v.AddConfigPath(path)
	}
	if err := v.ReadInConfig(); err != nil {
		log.Errorf("failed to read configuration file: %s", err)
		return nil
	}
	return v.Unmarshal(&Config)
}
