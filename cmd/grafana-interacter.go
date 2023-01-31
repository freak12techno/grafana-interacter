package main

import (
	"main/pkg"
	"main/pkg/app"
	"main/pkg/logger"

	"github.com/spf13/cobra"
)

func Execute(configPath string) {
	if configPath == "" {
		logger.GetDefaultLogger().Fatal().Msg("Cannot start without config")
	}

	config := pkg.LoadConfig(configPath)
	app := app.NewApp(config)

	app.Start()
}

func main() {
	var configPath string

	rootCmd := &cobra.Command{
		Use:  "grafana-interacter",
		Long: "A Telegram bot.",
		Run: func(cmd *cobra.Command, args []string) {
			Execute(configPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
