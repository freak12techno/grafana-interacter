package main

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleNewSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new silence query")

	silenceInfo, err := ParseSilenceOptions(c.Text(), c)
	if err != "" {
		return c.Reply(err)
	}

	_, silenceErr := a.Grafana.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	reply := fmt.Sprintf("<a href=\"%s\">Silence created.</a>", a.Grafana.RelativeLink("/alerting/silences"))
	return a.BotReply(c, reply)
}

func (a *App) HandleAlertmanagerNewSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new Alertmanager silence query")

	if !a.Alertmanager.Enabled() {
		return c.Reply("Alertmanager is disabled.")
	}

	silenceInfo, err := ParseSilenceOptions(c.Text(), c)
	if err != "" {
		return c.Reply(err)
	}

	silence, silenceErr := a.Alertmanager.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	reply := fmt.Sprintf("<a href=\"%s\">Silence created.</a>", a.Alertmanager.GetSilenceURL(silence))
	return a.BotReply(c, reply)
}
