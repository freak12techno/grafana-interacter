package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

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
