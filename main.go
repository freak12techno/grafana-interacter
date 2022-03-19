package main

import (
	"fmt"
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

	b.Handle("/dashboards", HandleListDashboards)
	b.Handle("/dashboard", HandleShowDashboard)
	b.Handle("/render", HandleRenderPanel)
	b.Start()

	log.Info().Msg("Telegram bot listening")
}

func HandleListDashboards(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboards query")

	dashboards, err := Grafana.GetAllDashboards()
	if err != nil {
		return c.Send(fmt.Sprintf("Error querying for dashboards: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Dashboards list</strong>:\n")
	for _, dashboard := range dashboards {
		sb.WriteString(fmt.Sprintf("- %s\n", Grafana.GetDashboardLink(dashboard)))
	}

	return c.Send(sb.String(), tele.ModeHTML)
}

func HandleShowDashboard(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboard query")

	args := strings.SplitAfterN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /render

	if len(args) != 1 {
		return c.Send("Usage: /dashboard <dashboard>")
	}

	dashboards, err := Grafana.GetAllDashboards()
	if err != nil {
		return c.Send(fmt.Sprintf("Error querying for dashboards: %s", err))
	}

	dashboard, found := FindDashboardByName(dashboards, args[0])
	if !found {
		return c.Send("Could not find dashboard. See /dashboards for dashboards list.")
	}

	dashboardEnriched, err := Grafana.GetDashboard(dashboard.UID)
	if err != nil {
		return c.Send(fmt.Sprintf("Could not get dashboard: %s", err))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<strong>Dashboard</strong> %s\n", Grafana.GetDashboardLink(*dashboard)))
	sb.WriteString("Panels:\n")
	for _, panel := range dashboardEnriched.Dashboard.Panels {
		sb.WriteString(fmt.Sprintf("- %s\n", Grafana.GetPanelLink(PanelStruct{
			DashboardURL: dashboard.URL,
			PanelID:      panel.ID,
			Name:         panel.Title,
		})))
	}

	return c.Send(sb.String(), tele.ModeHTML)
}

func HandleRenderPanel(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got render query")

	args := strings.SplitAfterN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /render

	if len(args) != 1 {
		return c.Send("Usage: /render <panel name>")
	}

	panels, err := Grafana.GetAllPanels()
	if err != nil {
		return c.Send(fmt.Sprintf("Error querying for panels: %s", err))
	}

	panel, found := FindPanelByName(panels, args[0])
	if !found {
		return c.Send("Could not find a panel. See /dashboards for dashboards list, and /dashboard <dashboard name> for its panels.")
	}

	image, err := Grafana.RenderPanel(panel)
	if err != nil {
		return c.Send(err)
	}

	defer image.Close()

	fileToSend := &tele.Photo{
		File:    tele.FromReader(image),
		Caption: fmt.Sprintf("Panel: %s", Grafana.GetPanelLink(*panel)),
	}
	return c.Send(fileToSend, tele.ModeHTML)
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
