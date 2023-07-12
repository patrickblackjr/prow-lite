package config

import (
	"io/ioutil"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var c Config

// Config is global object that holds all application level variables
type Config struct {
	Server    HTTPConfig       `yaml:"server"`
	GitHub    githubapp.Config `yaml:"github"`
	AppConfig AppConfiguration `yaml:"app_config"`
}

type HTTPConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type AppConfiguration struct{}

// ReadConfig unmarshals a YAML config file
func ReadConfig(path string) (*Config, error) {
	var c Config

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Errorf("failed to read configuration file: %s", err.Error())
		return nil, err
	}

	if err := yaml.Unmarshal(bytes, &c); err != nil {
		logrus.Errorf("failed to unmarshal configuration file: %s", err.Error())
		return nil, err
	}

	return &c, nil
}
