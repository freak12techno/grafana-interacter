package pkg

import (
	configPkg "main/pkg/config"
	"main/pkg/logger"
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) *configPkg.Config {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not read config file")
	}

	var config *configPkg.Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not unmarshal config file")
	}

	if err := defaults.Set(config); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set default settings")
	}

	if err := config.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Error validating config")
	}

	return config
}
