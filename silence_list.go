package main

import (
	"fmt"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListSilences(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got list silence query")

	silences, err := a.Grafana.GetSilences()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error listing silence: %s", err))
	}

	var sb strings.Builder
	sb.WriteString("<strong>Silences</strong>\n")

	for _, silence := range silences {
		sb.WriteString(silence.Serialize() + "\n")
		sb.WriteString(fmt.Sprintf("<a href=\"%s\">Link</a>\n\n", a.Grafana.RelativeLink("/alerting/silences")))
	}

	return a.BotReply(c, sb.String())
}

func (a *App) HandleAlertmanagerListSilences(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got Alertmanager list silence query")

	if !a.Alertmanager.Enabled() {
		return c.Reply("Alertmanager is disabled.")
	}

	silences, err := a.Alertmanager.GetSilences()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error listing silence: %s", err))
	}

	silences = Filter(silences, func(s Silence) bool {
		return s.EndsAt.After(time.Now())
	})

	var sb strings.Builder
	sb.WriteString("<strong>Silences</strong>\n")

	for _, silence := range silences {
		sb.WriteString(silence.Serialize() + "\n")
		sb.WriteString(fmt.Sprintf("<a href=\"%s\">Link</a>\n\n", a.Alertmanager.GetSilenceURL(silence)))
	}

	return a.BotReply(c, sb.String())
}
