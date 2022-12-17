package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListDatasources(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	datasources, err := a.Grafana.GetDatasources()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Datasources</strong>\n")
	for _, ds := range datasources {
		sb.WriteString(fmt.Sprintf("- %s\n", a.Grafana.GetDatasourceLink(ds)))
	}

	return a.BotReply(c, sb.String())
}
