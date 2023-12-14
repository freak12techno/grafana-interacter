package app

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleNewSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new silence query")

	silenceInfo, err := utils.ParseSilenceOptions(c.Text(), c)
	if err != "" {
		return c.Reply(err)
	}

	silenceResponse, silenceErr := a.Grafana.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	silence, silenceErr := a.Grafana.GetSilence(silenceResponse.SilenceID)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error getting created silence: %s", silenceErr))
	}

	template, renderErr := a.TemplateManager.Render("silences_create", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data: render.SilenceRender{
			Silence:       silence,
			AlertsPresent: false,
			Alerts:        []types.AlertmanagerAlert{},
		},
	})
	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silences_create template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	return a.BotReply(c, template)
}

func (a *App) HandleAlertmanagerNewSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got new Alertmanager silence query")

	if !a.Alertmanager.Enabled() {
		return c.Reply("Alertmanager is disabled.")
	}

	silenceInfo, err := utils.ParseSilenceOptions(c.Text(), c)
	if err != "" {
		return c.Reply(err)
	}

	silenceResponse, silenceErr := a.Alertmanager.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	silence, silenceErr := a.Alertmanager.GetSilence(silenceResponse.SilenceID)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error getting created silence: %s", silenceErr))
	}

	alerts, alertsErr := a.Alertmanager.GetSilenceMatchingAlerts(silence)
	if alertsErr != nil {
		return c.Reply(fmt.Sprintf("Error getting alerts for silence: %s", alertsErr))
	}

	template, renderErr := a.TemplateManager.Render("silences_create", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data: render.SilenceRender{
			Silence:       silence,
			AlertsPresent: alerts != nil,
			Alerts:        alerts,
		},
	})
	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silences_create template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	return a.BotReply(c, template)
}
