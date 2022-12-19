package main

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListFiringAlerts(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got firing alerts query")

	grafanaGroups, err := a.Grafana.GetGrafanaAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	prometheusGroups, err := a.Grafana.GetPrometheusAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	grafanaGroups = FilterFiringOrPendingAlertGroups(grafanaGroups)
	prometheusGroups = FilterFiringOrPendingAlertGroups(prometheusGroups)

	template, err := a.TemplateManager.Render("alerts_firing", RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data: AlertsListStruct{
			GrafanaGroups:    grafanaGroups,
			PrometheusGroups: prometheusGroups,
		},
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering alerts_firing template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}
