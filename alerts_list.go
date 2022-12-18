package main

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
)

type AlertsListStruct struct {
	GrafanaGroups    []GrafanaAlertGroup
	PrometheusGroups []GrafanaAlertGroup
}

func (a *App) HandleListAlerts(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	grafanaGroups, err := a.Grafana.GetGrafanaAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	prometheusGroups, err := a.Grafana.GetPrometheusAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	//var sb strings.Builder
	//if len(grafanaGroups) > 0 {
	//	sb.WriteString("<strong>Grafana alerts:</strong>\n")
	//	for _, group := range grafanaGroups {
	//		for _, rule := range group.Rules {
	//			sb.WriteString(rule.Serialize(group.Name))
	//		}
	//	}
	//} else {
	//	sb.WriteString("<strong>No Grafana alerts</strong>\n")
	//}
	//
	//if len(prometheusGroups) > 0 {
	//	sb.WriteString("<strong>Prometheus alerts:</strong>\n")
	//	for _, group := range prometheusGroups {
	//		for _, rule := range group.Rules {
	//			sb.WriteString(rule.Serialize(group.Name))
	//		}
	//	}
	//} else {
	//	sb.WriteString("<strong>No Prometheus alerts</strong>\n")
	//}

	template, err := a.TemplateManager.Render("alerts_list", RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data: AlertsListStruct{
			GrafanaGroups:    grafanaGroups,
			PrometheusGroups: prometheusGroups,
		},
	})

	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering alerts_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}
