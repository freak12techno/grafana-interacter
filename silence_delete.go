package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleDeleteSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new delete silence query")

	args := strings.SplitN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /unsilence

	if len(args) != 1 {
		return c.Reply("Usage: /unsilence <silence ID>")
	}

	silenceErr := a.Grafana.DeleteSilence(args[0])
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error deleting silence: %s", silenceErr))
	}

	return a.BotReply(c, "Silence deleted.")
}

func (a *App) HandleAlertmanagerDeleteSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new Alertmanager delete silence query")

	if !a.Alertmanager.Enabled() {
		return c.Reply("Alertmanager is disabled.")
	}

	args := strings.SplitN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /alertmanager_unsilence

	if len(args) != 1 {
		return c.Reply("Usage: /alertmanager_unsilence <silence ID>")
	}

	silenceErr := a.Alertmanager.DeleteSilence(args[0])
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error deleting silence: %s", silenceErr))
	}

	return a.BotReply(c, "Silence deleted.")
}
