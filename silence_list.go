package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func HandleListSilences(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got list silence query")

	silences, err := Grafana.GetSilences()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error listing silence: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Silences</strong>\n")

	for _, silence := range silences {
		sb.WriteString(silence.Serialize() + "\n")
		sb.WriteString(fmt.Sprintf("<a href=\"%s\">Link</a>\n\n", Grafana.RelativeLink("/alerting/silences")))

	}

	return BotReply(c, sb.String())
}

func HandleAlertmanagerListSilences(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got Alertmanager list silence query")

	silences, err := Alertmanager.GetSilences()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error listing silence: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Silences</strong>\n")

	for _, silence := range silences {
		sb.WriteString(silence.Serialize() + "\n")
		sb.WriteString(fmt.Sprintf("<a href=\"%s\">Link</a>\n\n", Alertmanager.GetSilenceURL(silence)))
	}

	return BotReply(c, sb.String())
}
