package main

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/creasty/defaults"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
	"gopkg.in/yaml.v2"
)

var (
	ConfigPath   string
	Config       ConfigStruct
	Grafana      *GrafanaStruct
	Alertmanager *AlertmanagerStruct
)

var log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

var rootCmd = &cobra.Command{
	Use:  "grafana-interacter",
	Long: "A Telegram bot.",
	Run:  Execute,
}

func Execute(cmd *cobra.Command, args []string) {
	if ConfigPath == "" {
		log.Fatal().Msg("Cannot start without config")
	}

	yamlFile, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not read config file")
	}

	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not unmarshal config file")
	}

	defaults.Set(&Config)

	logLevel, err := zerolog.ParseLevel(Config.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse log level")
	}

	if Config.JSONOutput {
		log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(logLevel)

	Grafana = InitGrafana(&Config.Grafana, &log)
	Alertmanager = InitAlertmanager(&Config.Alertmanager, &log)

	b, err := tele.NewBot(tele.Settings{
		Token:   Config.Telegram.Token,
		Poller:  &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: HandleError,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start Telegram bot")
		return
	}

	if len(Config.Telegram.Admins) > 0 {
		log.Debug().Msg("Using admins whitelist")
		b.Use(middleware.Whitelist(Config.Telegram.Admins...))
	}

	b.Handle("/dashboards", HandleListDashboards)
	b.Handle("/dashboard", HandleShowDashboard)
	b.Handle("/render", HandleRenderPanel)
	b.Handle("/datasources", HandleListDatasources)
	b.Handle("/alerts", HandleListAlerts)
	b.Handle("/alert", HandleSingleAlert)
	b.Handle("/silences", HandleListSilences)
	b.Handle("/silence", HandleNewSilence)
	b.Handle("/unsilence", HandleDeleteSilence)
	b.Handle("/alertmanager_silences", HandleAlertmanagerListSilences)
	b.Handle("/alertmanager_silence", HandleAlertmanagerNewSilence)
	b.Handle("/alertmanager_unsilence", HandleAlertmanagerDeleteSilence)

	log.Info().Msg("Telegram bot listening")

	b.Start()
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
