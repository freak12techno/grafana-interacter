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

	if len(args) <= 1 {
		return c.Reply("Usage: /unsilence <silence ID or labels>")
	}

	silences, silencesFetchErr := a.Grafana.GetSilences()
	if silencesFetchErr != nil {
		return c.Reply(fmt.Sprintf("Error getting silence to delete: %s", silencesFetchErr))
	}

	silence, found, err := silences.FindByNameOrMatchers(args[1])
	if !found {
		return c.Reply(fmt.Sprintf("Silence is not found by ID or matchers: %s", args[0]))
	}

	if err != "" {
		return c.Reply(fmt.Sprintf("Error getting silence by ID or matchers: %s", err))
	}

	if silence.Status.State == "expired" {
		return c.Reply("Silence is already deleted!")
	}

	silenceErr := a.Grafana.DeleteSilence(silence.ID)
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

	if len(args) <= 1 {
		return c.Reply("Usage: /alertmanager_unsilence <silence ID or labels>")
	}

	silences, silencesFetchErr := a.Alertmanager.GetSilences()
	if silencesFetchErr != nil {
		return c.Reply(fmt.Sprintf("Error getting silence to delete: %s", silencesFetchErr))
	}

	silence, found, err := silences.FindByNameOrMatchers(args[1])
	if !found {
		return c.Reply(fmt.Sprintf("Silence is not found by ID or matchers: %s", args[0]))
	}

	if err != "" {
		return c.Reply(fmt.Sprintf("Error getting silence by ID or matchers: %s", err))
	}

	if silence.Status.State == "expired" {
		return c.Reply("Silence is already deleted!")
	}

	silenceErr := a.Alertmanager.DeleteSilence(silence.ID)
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
