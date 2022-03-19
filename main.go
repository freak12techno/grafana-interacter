package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

	from := time.Now().Unix() * 1000
	to := time.Now().Add(-30*time.Minute).Unix() * 1000

	dashboardID := panel.DashboardID
	panelID := panel.PanelID

	url := fmt.Sprintf(
		"%s/render/d-solo/%s/dashboard?orgId=1&from=%d&to=%d&panelId=%s&width=1000&height=500&tz=Europe/Moscow",
		Config.GrafanaURL,
		dashboardID,
		from,
		to,
		panelID,
	)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	if Config.Auth.User != "" && Config.Auth.Password != "" {
		log.Trace().Msg("Using basic auth")
		req.SetBasicAuth(Config.Auth.User, Config.Auth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return c.Send(fmt.Sprintf("Could not query dashboard: %s", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return c.Send(fmt.Sprintf("Could not fetch rendered image. Status code: %d", resp.StatusCode))
	}

	fileToSend := &tele.Photo{File: tele.FromReader(resp.Body)}
	return c.Send(fileToSend)
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
