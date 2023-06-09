package app

import (
	"fmt"
	"main/pkg/types/render"
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

	silence, silenceFetchErr := a.Grafana.GetSilence(args[0])
	if silenceFetchErr != nil {
		return c.Reply(fmt.Sprintf("Error getting silence to delete: %s", silenceFetchErr))
	}

	silenceErr := a.Grafana.DeleteSilence(args[0])
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error deleting silence: %s", silenceErr))
	}

	template, renderErr := a.TemplateManager.Render("silences_delete", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         silence,
	})

	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silences_delete template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	return a.BotReply(c, template)
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

	silence, silenceFetchErr := a.Alertmanager.GetSilence(args[0])
	if silenceFetchErr != nil {
		return c.Reply(fmt.Sprintf("Error getting silence to delete: %s", silenceFetchErr))
	}

	silenceErr := a.Alertmanager.DeleteSilence(args[0])
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error deleting silence: %s", silenceErr))
	}

	template, renderErr := a.TemplateManager.Render("silences_delete", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         silence,
	})

	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silences_delete template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	return a.BotReply(c, template)
}
