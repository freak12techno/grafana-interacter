package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/yaml.v2"
)

var (
	ConfigPath string
	Config     ConfigStruct
	Grafana    *GrafanaStruct
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

	Grafana = InitGrafana(Config.GrafanaURL, &Config.Auth, &log)

	b, err := tele.NewBot(tele.Settings{
		Token:  Config.TelegramToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start Telegram bot")
		return
	}

	b.Handle("/dashboard", HandleDashboard)
	b.Start()

	log.Info().Msg("Telegram bot listening")
}

func HandleDashboard(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboard query")

	args := strings.SplitAfterN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /dashboard

	if len(args) != 1 {
		return c.Send("Usage: /dashboard <dashboard ID>")
	}

	var panel *PanelStruct
	for _, p := range Config.Panels {
		log.Trace().Str("name", p.Name).Msg("Iterating over panel")
		if p.Name == args[0] {
			panel = &p
			break
		}
	}

	if panel == nil {
		var sb strings.Builder
		sb.WriteString("Could not find this panel. Available panels are:\n")
		for _, p := range Config.Panels {
			sb.WriteString("- " + p.Name + "\n")
		}

		return c.Send(sb.String())
	}

	image, err := Grafana.RenderPanel(panel)
	if err != nil {
		return c.Send(err)
	}

	defer image.Close()

	fileToSend := &tele.Photo{File: tele.FromReader(image)}
	return c.Send(fileToSend)
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
