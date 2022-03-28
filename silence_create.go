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

	silenceErr := Grafana.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	return BotReply(c, "Silence created.")
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

	silenceErr := Alertmanager.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	return BotReply(c, "Silence created.")
}
