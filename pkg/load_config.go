package pkg

import (
	configPkg "main/pkg/config"
	"main/pkg/fs"
	"main/pkg/logger"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

func LoadConfig(filesystem fs.FS, path string) *configPkg.Config {
	yamlFile, err := filesystem.ReadFile(path)
	if err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not read config file")
	}

	var config *configPkg.Config
	if parseErr := yaml.Unmarshal(yamlFile, &config); parseErr != nil {
		logger.GetDefaultLogger().Panic().Err(parseErr).Msg("Could not unmarshal config file")
	}

	defaults.MustSet(config)
	return config
}
