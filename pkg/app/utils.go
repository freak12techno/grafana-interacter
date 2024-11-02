package app

import (
	"strings"

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
