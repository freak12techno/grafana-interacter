package pkg

import (
	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
	"main/pkg/config"
	"main/pkg/logger"
	"os"
)

func LoadConfig(path string) *config.Config {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not read config file")
	}

	var config *config.Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not unmarshal config file")
	}

	defaults.Set(&config)

	return config
}
