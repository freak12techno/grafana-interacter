package main

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func HandleNewSilence(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new silence query")

	silenceInfo, err := ParseSilenceOptions(c.Text(), c)
	if err != "" {
		return c.Reply(err)
	}

	_, silenceErr := Grafana.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	reply := fmt.Sprintf("<a href=\"%s\">Silence created.</a>", Grafana.RelativeLink("/alerting/silences"))
	return BotReply(c, reply)
}

func HandleAlertmanagerNewSilence(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new Alertmanager silence query")

	silenceInfo, err := ParseSilenceOptions(c.Text(), c)
	if err != "" {
		return c.Reply(err)
	}

	silence, silenceErr := Alertmanager.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	reply := fmt.Sprintf("<a href=\"%s\">Silence created.</a>", Alertmanager.GetSilenceURL(silence))
	return BotReply(c, reply)
}
