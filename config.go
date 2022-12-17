package main

import (
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Log          LogConfig          `yaml:"log"`
	Telegram     TelegramConfig     `yaml:"telegram" default:""`
	Grafana      GrafanaConfig      `yaml:"grafana"`
	Alertmanager AlertmanagerConfig `yaml:"alertmanager"`
}

type LogConfig struct {
	LogLevel   string `yaml:"level" default:"info"`
	JSONOutput bool   `yaml:"json" default:"false"`
}

type TelegramConfig struct {
	Token  string  `yaml:"token"`
	Admins []int64 `yaml:"admins"`
}

type GrafanaConfig struct {
	URL      string `yaml:"url" default:"http://localhost:3000"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Timezone string `yaml:"timezone" default:"Europe/Moscow"`
}

type AlertmanagerConfig struct {
	URL      string `yaml:"url" default:"http://localhost:9093"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Timezone string `yaml:"timezone" default:"Europe/Moscow"`
}

func LoadConfig(path string) *Config {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not read config file")
	}

	var config *Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not unmarshal config file")
	}

	defaults.Set(&config)

	return config
}
