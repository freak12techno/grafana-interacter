package app

import (
	"fmt"
	"main/pkg/alert_source"
	"main/pkg/silence_manager"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleNewSilenceViaCommand(silenceManager silence_manager.SilenceManager) func(c tele.Context) error {
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

func (a *App) HandlePrepareNewSilenceFromCallback(
	silenceManager silence_manager.SilenceManager,
	alertSource alert_source.AlertSource,
) func(c tele.Context) error {
	return func(c tele.Context) error {
		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("silence_manager", silenceManager.Name()).
			Str("alert_source", alertSource.Name()).
			Msg("Got new prepare silence callback via button")

		callback := c.Callback()
		a.RemoveKeyboardItemByCallback(c, callback)

		groups, err := alertSource.GetAlertingRules()
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
			Grafana: a.Grafana,
			Data:    matchers,
		})
		if renderErr != nil {
			a.Logger.Error().Err(renderErr).Msg("Error rendering silence_prepare_create template")
			return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
		}

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		mutesDurations := silenceManager.GetMutesDurations()
		rows := make([]tele.Row, len(mutesDurations))

		for index, mute := range mutesDurations {
			rows[index] = menu.Row(menu.Data(
				fmt.Sprintf("⌛ Silence for %s", mute),
				silenceManager.Prefixes().Silence,
				mute+" "+callback.Data,
			))
		}

		menu.Inline(rows...)
		return a.BotReply(c, template, menu)
	}
}

func (a *App) HandleCallbackNewSilence(
	silenceManager silence_manager.SilenceManager,
	alertSource alert_source.AlertSource,
) func(c tele.Context) error {
	return func(c tele.Context) error {
		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("silence_manager", silenceManager.Name()).
			Str("alert_source", alertSource.Name()).
			Msg("Got new create silence callback via button")

		callback := c.Callback()
		a.RemoveKeyboardItemByCallback(c, callback)

		dataSplit := strings.SplitN(callback.Data, " ", 2)
		if len(dataSplit) != 2 {
			return c.Reply("Invalid callback provided!")
		}

		alertHashToMute := dataSplit[1]

		groups, err := alertSource.GetAlertingRules()
		if err != nil {
			return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
		}

		silenceInfo, err := a.GenerateSilenceForAlert(c, groups, alertHashToMute, dataSplit[0])
		if err != nil {
			return c.Reply(err.Error())
		}

		return a.HandleNewSilenceGeneric(c, silenceManager, silenceInfo)
	}
}

func (a *App) HandleNewSilenceGeneric(
	c tele.Context,
	silenceManager silence_manager.SilenceManager,
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

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	menu.Inline(menu.Row(menu.Data(
		"❌Unsilence",
		silenceManager.Prefixes().Unsilence,
		silence.ID,
	)))

	return a.ReplyRenderWithMarkup(c, "silences_create", render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.SilenceWithAlerts{
			Silence:       silence,
			AlertsPresent: alerts != nil,
			Alerts:        alerts,
		},
	}, menu)
}
