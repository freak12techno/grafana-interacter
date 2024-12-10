package app

import (
	"fmt"
	"main/pkg/alert_source"
	"main/pkg/constants"
	"main/pkg/silence_manager"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils"
	"strings"
	"time"

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
			return c.Reply(err)
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
			Str("callback", c.Callback().Data).
			Msg("Got new prepare silence callback via button")

		callback := c.Callback()
		callbackSplit := strings.SplitN(callback.Data, " ", 2)
		a.RemoveKeyboardItemByCallback(c, callback)

		labels, found := a.Cache.Get(callbackSplit[0])
		if !found {
			return c.Reply("Alert was not found!")
		}

		matchers := types.QueryMatcherFromKeyValueString(labels)
		matchers.Sort()

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		mutesDurations := silenceManager.GetMutesDurations()
		rows := make([]tele.Row, 0)

		if len(matchers) > 1 {
			for _, matcher := range matchers {
				matchersWithoutKey := matchers.WithoutKey(matcher.Key)
				cacheKey := a.Cache.Set(matchersWithoutKey.GetHash(), matchersWithoutKey.ToQueryString())

				rows = append(rows, menu.Row(menu.Data(
					fmt.Sprintf("❌ Remove %s", matcher.Serialize()),
					silenceManager.Prefixes().PrepareSilence,
					cacheKey+" 1", // to update the message instead of editing
				)))
			}
		}

		for _, mute := range mutesDurations {
			rows = append(rows, menu.Row(menu.Data(
				fmt.Sprintf("⌛ Silence for %s", mute),
				silenceManager.Prefixes().Silence,
				mute+" "+callbackSplit[0],
			)))
		}

		menu.Inline(rows...)

		if len(callbackSplit) > 1 {
			return a.EditRender(c, "silence_prepare_create", render.RenderStruct{
				Grafana: a.Grafana,
				Data:    matchers,
			}, menu)
		}

		return a.ReplyRender(c, "silence_prepare_create", render.RenderStruct{
			Grafana: a.Grafana,
			Data:    matchers,
		}, menu)
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

		durationRaw := dataSplit[0]
		alertLabelsRaw := dataSplit[1]

		duration, err := time.ParseDuration(durationRaw)
		if err != nil {
			return c.Reply("Invalid duration provided!")
		}

		labels, found := a.Cache.Get(alertLabelsRaw)
		if !found {
			return c.Reply("Alert was not found!")
		}

		matchers := types.QueryMatcherFromKeyValueString(labels)
		silenceInfo, _ := utils.ParseSilenceWithDuration("callback", matchers, c.Sender().FirstName, duration)

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
		"✅Confirm",
		constants.ClearKeyboardPrefix,
	)), menu.Row(menu.Data(
		"❌Unsilence",
		silenceManager.Prefixes().Unsilence,
		silence.ID+" 1",
	)))

	return a.ReplyRender(c, "silences_create", render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.SilenceWithAlerts{
			Silence:       silence,
			AlertsPresent: alerts != nil,
			Alerts:        alerts,
		},
	}, menu)
}
