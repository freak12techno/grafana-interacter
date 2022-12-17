package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	tele "gopkg.in/telebot.v3"
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

func HandleError(err error, c tele.Context) {
	log.Error().Err(err).Msg("Telebot error")
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
