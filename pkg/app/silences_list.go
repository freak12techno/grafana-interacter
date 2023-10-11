package app

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils/generic"
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

	silences = generic.Filter(silences, func(s types.Silence) bool {
		return s.EndsAt.After(time.Now())
	})

	template, err := a.TemplateManager.Render("silences_list", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         silences,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering silences_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
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

	silences = generic.Filter(silences, func(s types.Silence) bool {
		return s.EndsAt.After(time.Now())
	})

	template, err := a.TemplateManager.Render("silences_list", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         silences,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering silences_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}
