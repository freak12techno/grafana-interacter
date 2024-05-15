package main

import (
	"main/pkg"
	"main/pkg/app"
	"main/pkg/logger"

	"github.com/spf13/cobra"
)

var version = "unknown"

func ExecuteMain(configPath string) {
	if configPath == "" {
		logger.GetDefaultLogger().Fatal().Msg("Cannot start without config")
	}

	config := pkg.LoadConfig(configPath)
	if err := config.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Error validating config")
	}

	newApp := app.NewApp(config, version)
	newApp.Start()
}

func ExecuteValidateConfig(configPath string) {
	config := pkg.LoadConfig(configPath)

	if err := config.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Error validating config")
	}

	logger.GetDefaultLogger().Info().Msg("Provided config is valid.")
}

func main() {
	var configPath string

	rootCmd := &cobra.Command{
		Use:     "grafana-interacter --config [config path]",
		Long:    "A Telegram bot to interact with your Grafana, Prometheus and Alertmanager instances.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteMain(configPath)
		},
	}

	validateConfigCmd := &cobra.Command{
		Use:     "validate-config --config [config path]",
		Long:    "Validate config.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteValidateConfig(configPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	validateConfigCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path")
	if err := validateConfigCmd.MarkPersistentFlagRequired("config"); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	rootCmd.AddCommand(validateConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
