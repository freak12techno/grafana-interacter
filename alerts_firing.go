package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func HandleListFiringAlerts(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got firing alerts query")

	grafanaGroups, err := Grafana.GetGrafanaAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	prometheusGroups, err := Grafana.GetPrometheusAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	grafanaGroups = FilterFiringOrPendingAlertGroups(grafanaGroups)
	prometheusGroups = FilterFiringOrPendingAlertGroups(prometheusGroups)

	var sb strings.Builder
	if len(grafanaGroups) > 0 {
		sb.WriteString("<strong>Grafana alerts:</strong>\n")
		for _, group := range grafanaGroups {
			for _, rule := range group.Rules {
				sb.WriteString(rule.SerializeFull(group.Name))
			}
		}
	} else {
		sb.WriteString("<strong>No Grafana alerts</strong>\n")
	}

	if len(prometheusGroups) > 0 {
		sb.WriteString("<strong>Prometheus alerts:</strong>\n")
		for _, group := range prometheusGroups {
			for _, rule := range group.Rules {
				sb.WriteString(rule.SerializeFull(group.Name))
			}
		}
	} else {
		sb.WriteString("<strong>No Prometheus alerts</strong>\n")
	}

	return BotReply(c, sb.String())
}
