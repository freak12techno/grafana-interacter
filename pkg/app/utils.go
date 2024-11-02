package app

import (
	"errors"
	"fmt"
	"main/pkg/types"
	"main/pkg/utils"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (a *App) RemoveKeyboardItemByCallback(c tele.Context, callback *tele.Callback) {
	if callback.Message != nil && callback.Message.ReplyMarkup != nil {
		for rowIndex, row := range callback.Message.ReplyMarkup.InlineKeyboard {
			for itemIndex, item := range row {
				split := strings.SplitN(item.Data, "|", 2)
				if len(split) != 2 {
					continue
				}

				if split[1] == callback.Data {
					callback.Message.ReplyMarkup.InlineKeyboard[rowIndex] = append(
						callback.Message.ReplyMarkup.InlineKeyboard[rowIndex][:itemIndex],
						callback.Message.ReplyMarkup.InlineKeyboard[rowIndex][itemIndex+1:]...,
					)
				}
			}
		}

		if _, err := a.Bot.EditReplyMarkup(
			callback.Message,
			callback.Message.ReplyMarkup,
		); err != nil {
			a.Logger.Error().
				Str("sender", c.Sender().Username).
				Err(err).
				Msg("Error updating message when editing a callback")
		}
	}
}

func (a *App) GenerateSilenceForAlert(
	c tele.Context,
	groups types.GrafanaAlertGroups,
	alertHashToMute string,
	durationRaw string,
) (*types.Silence, error) {
	duration, err := time.ParseDuration(durationRaw)
	if err != nil {
		return nil, fmt.Errorf("Invalid duration provided!")
	}

	groups = groups.FilterFiringOrPendingAlertGroups()

	for _, group := range groups {
		for _, rule := range group.Rules {
			for _, alert := range rule.Alerts {
				alertHash := alert.GetCallbackHash()
				if alertHash != alertHashToMute {
					continue
				}

				matchers := types.QueryMatcherFromKeyValueMap(alert.Labels)
				silenceInfo, silenceErr := utils.ParseSilenceWithDuration("callback", matchers, c.Sender().FirstName, duration)
				if silenceErr != "" {
					return nil, fmt.Errorf("Error parsing silence option: %s\n", silenceErr)
				}

				return silenceInfo, nil
			}
		}
	}

	return nil, errors.New("Alert was not found!")
}
