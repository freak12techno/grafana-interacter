package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListDashboards(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboards query")

	dashboards, err := a.Grafana.GetAllDashboards()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for dashboards: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Dashboards list</strong>:\n")
	for _, dashboard := range dashboards {
		sb.WriteString(fmt.Sprintf("- %s\n", a.Grafana.GetDashboardLink(dashboard)))
	}

	return a.BotReply(c, sb.String())
}
