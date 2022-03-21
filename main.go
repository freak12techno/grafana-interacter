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
	b.Handle("/datasources", HandleListDatasources)
	b.Handle("/alerts", HandleListAlerts)
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

	opts, valid := ParseRenderOptions(c.Text())

	if !valid {
		return c.Send("Usage: /render <opts> <panel name>")
	}

	panels, err := Grafana.GetAllPanels()
	if err != nil {
		return c.Send(fmt.Sprintf("Error querying for panels: %s", err))
	}

	panel, found := FindPanelByName(panels, opts.Query)
	if !found {
		return c.Send("Could not find a panel. See /dashboards for dashboards list, and /dashboard <dashboard name> for its panels.")
	}

	image, err := Grafana.RenderPanel(panel, opts.Params)
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

func HandleListDatasources(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	datasources, err := Grafana.GetDatasources()
	if err != nil {
		return c.Send(fmt.Sprintf("Error querying alerts: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Datasources</strong>\n")
	for _, ds := range datasources {
		sb.WriteString(fmt.Sprintf("- %s\n", Grafana.GetDatasourceLink(ds)))
	}

	return c.Send(sb.String(), tele.ModeHTML)
}

func HandleListAlerts(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	grafanaGroups, err := Grafana.GetGrafanaAlertingRules()
	if err != nil {
		return c.Send(fmt.Sprintf("Error querying alerts: %s", err))
	}

	prometheusGroups, err := Grafana.GetPrometheusAlertingRules()
	if err != nil {
		return c.Send(fmt.Sprintf("Error querying alerts: %s", err))
	}

	var sb strings.Builder
	if len(grafanaGroups) > 0 {
		sb.WriteString("<strong>Grafana alerts</strong>\n")
		for _, group := range grafanaGroups {
			for _, rule := range group.Rules {
				switch rule.State {
				case "inactive":
					sb.WriteString(fmt.Sprintf("- 🟢 %s -> %s\n", group.Name, rule.Name))
				case "pending":
					sb.WriteString(fmt.Sprintf("- 🟡 %s -> %s\n", group.Name, rule.Name))
				case "firing":
					sb.WriteString(fmt.Sprintf("- 🔴 %s -> %s\n", group.Name, rule.Name))
				default:
					sb.WriteString(fmt.Sprintf("- [%s] %s -> %s\n", rule.State, group.Name, rule.Name))
				}

			}
		}
	} else {
		sb.WriteString("<strong>No Grafana alerts</strong>\n")
	}

	if len(prometheusGroups) > 0 {
		sb.WriteString("<strong>Prometheus alerts</strong>\n")
		for _, group := range prometheusGroups {
			for _, rule := range group.Rules {
				switch rule.State {
				case "inactive":
					sb.WriteString(fmt.Sprintf("- 🟢 %s -> %s\n", group.Name, rule.Name))
				case "pending":
					sb.WriteString(fmt.Sprintf("- 🟡 %s -> %s\n", group.Name, rule.Name))
				case "firing":
					sb.WriteString(fmt.Sprintf("- 🔴 %s -> %s\n", group.Name, rule.Name))
				default:
					sb.WriteString(fmt.Sprintf("- [%s] %s -> %s\n", rule.State, group.Name, rule.Name))
				}

			}
		}
	} else {
		sb.WriteString("<strong>No Prometheus alerts</strong>\n")
	}

	return c.Send(sb.String(), tele.ModeHTML)
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
