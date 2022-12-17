package main

import (
	"github.com/spf13/cobra"
)

var (
	ConfigPath string
)

var rootCmd = &cobra.Command{
	Use:  "grafana-interacter",
	Long: "A Telegram bot.",
	Run:  Execute,
}

func Execute(cmd *cobra.Command, args []string) {
	if ConfigPath == "" {
		GetDefaultLogger().Fatal().Msg("Cannot start without config")
	}

	config := LoadConfig(ConfigPath)
	app := NewApp(config)

	app.Start()
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
