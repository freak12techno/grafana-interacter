package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func HandleAlertmanagerDeleteSilence(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new Alertmanager delete silence query")

	args := strings.SplitN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /alertmanager_unsilence

	if len(args) != 1 {
		return c.Reply("Usage: /alertmanager_unsilence <silence ID>")
	}

	silenceErr := Alertmanager.DeleteSilence(args[0])
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error deleting silence: %s", silenceErr))
	}

	return BotReply(c, "Silence deleted.")
}
