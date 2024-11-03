package app

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleNewSilenceViaCommand(silenceManager types.SilenceManager) func(c tele.Context) error {
	return func(c tele.Context) error {
		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Str("silence_manager", silenceManager.Name()).
			Msg("Got new silence query")

		if !silenceManager.Enabled() {
			return c.Reply(silenceManager.Name() + " is disabled.")
		}

		silenceInfo, err := utils.ParseSilenceFromCommand(c.Text(), c.Sender().FirstName)
		if err != "" {
			return c.Reply(fmt.Sprintf("Error parsing silence option: %s\n", err))
		}

		return a.HandleNewSilenceGeneric(c, silenceManager, silenceInfo)
	}
}

func (a *App) HandleGrafanaPrepareNewSilenceFromCallback(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Msg("Got new prepare Grafana silence callback via button")

	callback := c.Callback()
	a.RemoveKeyboardItemByCallback(c, callback)

	groups, err := a.Grafana.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	groups = groups.FilterFiringOrPendingAlertGroups()
	labels, found := groups.FindLabelsByHash(callback.Data)
	if !found {
		return c.Reply("Alert was not found!")
	}

	matchers := types.QueryMatcherFromKeyValueMap(labels)
	template, renderErr := a.TemplateManager.Render("silence_prepare_create", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         matchers,
	})
	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silence_prepare_create template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	rows := make([]tele.Row, len(a.Config.Grafana.MutesDurations))

	for index, mute := range a.Config.Grafana.MutesDurations {
		rows[index] = menu.Row(menu.Data(
			fmt.Sprintf("⌛ Silence for %s", mute),
			constants.GrafanaSilencePrefix,
			mute+" "+callback.Data,
		))
	}

	menu.Inline(rows...)
	return a.BotReply(c, template, menu)
}

func (a *App) HandleAlertmanagerPrepareNewSilenceFromCallback(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Msg("Got new prepare Alertmanager silence callback via button")

	callback := c.Callback()
	a.RemoveKeyboardItemByCallback(c, callback)

	groups, err := a.Prometheus.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	groups = groups.FilterFiringOrPendingAlertGroups()
	labels, found := groups.FindLabelsByHash(callback.Data)
	if !found {
		return c.Reply("Alert was not found!")
	}

	matchers := types.QueryMatcherFromKeyValueMap(labels)
	template, renderErr := a.TemplateManager.Render("silence_prepare_create", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         matchers,
	})
	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silence_prepare_create template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	rows := make([]tele.Row, len(a.Config.Alertmanager.MutesDurations))

	for index, mute := range a.Config.Alertmanager.MutesDurations {
		rows[index] = menu.Row(menu.Data(
			fmt.Sprintf("⌛ Silence for %s", mute),
			constants.AlertmanagerSilencePrefix,
			mute+" "+callback.Data,
		))
	}

	menu.Inline(rows...)
	return a.BotReply(c, template, menu)
}

func (a *App) HandleGrafanaCallbackNewSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Msg("Got new create Grafana silence callback via button")

	callback := c.Callback()
	a.RemoveKeyboardItemByCallback(c, callback)

	dataSplit := strings.SplitN(callback.Data, " ", 2)
	if len(dataSplit) != 2 {
		return c.Reply("Invalid callback provided!")
	}

	alertHashToMute := dataSplit[1]

	groups, err := a.Grafana.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	silenceInfo, err := a.GenerateSilenceForAlert(c, groups, alertHashToMute, dataSplit[0])
	if err != nil {
		return c.Reply(err.Error())
	}

	return a.HandleNewSilenceGeneric(c, a.Grafana, silenceInfo)
}

func (a *App) HandleAlertmanagerCallbackNewSilence(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Msg("Got new create Alertmanager silence callback via button")

	callback := c.Callback()
	a.RemoveKeyboardItemByCallback(c, callback)

	dataSplit := strings.SplitN(callback.Data, " ", 2)
	if len(dataSplit) != 2 {
		return c.Reply("Invalid callback provided!")
	}

	alertHashToMute := dataSplit[1]

	groups, err := a.Prometheus.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	silenceInfo, err := a.GenerateSilenceForAlert(c, groups, alertHashToMute, dataSplit[0])
	if err != nil {
		return c.Reply(err.Error())
	}

	return a.HandleNewSilenceGeneric(c, a.Alertmanager, silenceInfo)
}

func (a *App) HandleNewSilenceGeneric(
	c tele.Context,
	silenceManager types.SilenceManager,
	silenceInfo *types.Silence,
) error {
	silenceResponse, silenceErr := silenceManager.CreateSilence(*silenceInfo)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error creating silence: %s", silenceErr))
	}

	silence, silenceErr := silenceManager.GetSilence(silenceResponse.SilenceID)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error getting created silence: %s", silenceErr))
	}

	alerts, alertsErr := silenceManager.GetSilenceMatchingAlerts(silence)
	if alertsErr != nil {
		return c.Reply(fmt.Sprintf("Error getting alerts for silence: %s", alertsErr))
	}

	template, renderErr := a.TemplateManager.Render("silences_create", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data: types.SilenceWithAlerts{
			Silence:       silence,
			AlertsPresent: alerts != nil,
			Alerts:        alerts,
		},
	})
	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silences_create template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	menu.Inline(menu.Row(menu.Data(
		"❌Unsilence",
		silenceManager.GetUnsilencePrefix(),
		silence.ID,
	)))

	return a.BotReply(c, template, menu)
}
