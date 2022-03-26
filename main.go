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
	"gopkg.in/telebot.v3/middleware"
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

	Grafana = InitGrafana(&Config.Grafana, &log)

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
	b.Handle("/silence", HandleNewSilence)

	log.Info().Msg("Telegram bot listening")

	b.Start()
}

func HandleError(err error, c tele.Context) {
	log.Error().Err(err).Msg("Telebot error")
}

func HandleListDashboards(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboards query")

	dashboards, err := Grafana.GetAllDashboards()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for dashboards: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Dashboards list</strong>:\n")
	for _, dashboard := range dashboards {
		sb.WriteString(fmt.Sprintf("- %s\n", Grafana.GetDashboardLink(dashboard)))
	}

	return BotReply(c, sb.String())
}

func HandleShowDashboard(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboard query")

	args := strings.SplitN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /render

	if len(args) != 1 {
		return c.Reply("Usage: /dashboard <dashboard>")
	}

	dashboards, err := Grafana.GetAllDashboards()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for dashboards: %s", err))
	}

	dashboard, found := FindDashboardByName(dashboards, args[0])
	if !found {
		return c.Reply("Could not find dashboard. See /dashboards for dashboards list.")
	}

	dashboardEnriched, err := Grafana.GetDashboard(dashboard.UID)
	if err != nil {
		return c.Reply(fmt.Sprintf("Could not get dashboard: %s", err))
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

	return BotReply(c, sb.String())
}

func HandleRenderPanel(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got render query")

	opts, valid := ParseRenderOptions(c.Text())

	if !valid {
		return c.Reply("Usage: /render <opts> <panel name>")
	}

	panels, err := Grafana.GetAllPanels()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for panels: %s", err))
	}

	panel, found := FindPanelByName(panels, opts.Query)
	if !found {
		return c.Reply("Could not find a panel. See /dashboards for dashboards list, and /dashboard <dashboard name> for its panels.")
	}

	image, err := Grafana.RenderPanel(panel, opts.Params)
	if err != nil {
		return c.Reply(err)
	}

	defer image.Close()

	fileToSend := &tele.Photo{
		File:    tele.FromReader(image),
		Caption: fmt.Sprintf("Panel: %s", Grafana.GetPanelLink(*panel)),
	}
	return c.Reply(fileToSend, tele.ModeHTML)
}

func HandleListDatasources(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	datasources, err := Grafana.GetDatasources()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Datasources</strong>\n")
	for _, ds := range datasources {
		sb.WriteString(fmt.Sprintf("- %s\n", Grafana.GetDatasourceLink(ds)))
	}

	return BotReply(c, sb.String())
}

func HandleListAlerts(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	grafanaGroups, err := Grafana.GetGrafanaAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	prometheusGroups, err := Grafana.GetPrometheusAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	var sb strings.Builder
	if len(grafanaGroups) > 0 {
		sb.WriteString("<strong>Grafana alerts</strong>\n")
		for _, group := range grafanaGroups {
			for _, rule := range group.Rules {
				sb.WriteString(rule.Serialize(group.Name))
			}
		}
	} else {
		sb.WriteString("<strong>No Grafana alerts</strong>\n")
	}

	if len(prometheusGroups) > 0 {
		sb.WriteString("<strong>Prometheus alerts</strong>\n")
		for _, group := range prometheusGroups {
			for _, rule := range group.Rules {
				sb.WriteString(rule.Serialize(group.Name))
			}
		}
	} else {
		sb.WriteString("<strong>No Prometheus alerts</strong>\n")
	}

	return BotReply(c, sb.String())
}

func HandleSingleAlert(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got single alert query")

	args := strings.SplitN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /alert

	if len(args) != 1 {
		return c.Reply("Usage: /alert <alert name>")
	}

	rules, err := Grafana.GetAllAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	rule, found := FindAlertRuleByName(rules, args[0])
	if !found {
		return c.Reply("Could not find alert. See /alert for alerting rules.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<strong>Alert rule: </strong>%s\n", rule.Name))
	sb.WriteString("<strong>Alerts: </strong>\n")

	for _, alert := range rule.Alerts {
		sb.WriteString(alert.Serialize())
	}

	return BotReply(c, sb.String())
}

func HandleNewSilence(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new silence query")

	silenceInfo, err := ParseSilenceOptions(c.Text(), c)
	if err != "" {
		return c.Reply(err)
	}

	silenceErr := Grafana.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	return c.Reply("Silence created.")
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
