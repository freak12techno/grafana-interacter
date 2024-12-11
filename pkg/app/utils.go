package app

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) RemoveKeyboardItemByCallback(c tele.Context, callback *tele.Callback) {
	if callback.Message == nil || callback.Message.ReplyMarkup == nil {
		return
	}

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

func (a *App) ClearKeyboard(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Msg("Got new clear keyboard query")

	callback := c.Callback()
	if callback.Message == nil || callback.Message.ReplyMarkup == nil {
		return nil
	}

	for _, row := range callback.Message.ReplyMarkup.InlineKeyboard {
		for _, item := range row {
			split1 := strings.SplitN(item.Data, "|", 2)
			if len(split1) != 2 {
				continue
			}

			split2 := strings.SplitN(split1[1], " ", 2)
			if len(split2) < 1 {
				continue
			}

			a.Cache.Delete(split2[0])
		}
	}

	fmt.Printf("cache: %+v\n", a.Cache)

	if _, err := a.Bot.EditReplyMarkup(
		callback.Message,
		nil,
	); err != nil {
		a.Logger.Error().
			Str("sender", c.Sender().Username).
			Err(err).
			Msg("Error clearing keyboard when editing a callback")
		return err
	}

	return nil
}

func (a *App) GetAllAlertingRules() (types.GrafanaAlertGroups, error) {
	rules := make(types.GrafanaAlertGroups, 0)

	for _, alertSource := range a.AlertSourcesWithSilenceManager {
		alertSourceRules, err := alertSource.AlertSource.GetAlertingRules()
		if err != nil {
			return nil, err
		}

		rules = append(rules, alertSourceRules...)
	}

	return rules, nil
}

func (a *App) ReplyRender(
	c tele.Context,
	templateName string,
	renderStruct render.RenderStruct,
	opts ...interface{},
) error {
	template, err := a.TemplateManager.Render(templateName, renderStruct)
	if err != nil {
		a.Logger.Error().Str("template", templateName).Err(err).Msg("Error rendering template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template, opts...)
}

func (a *App) EditRender(
	c tele.Context,
	templateName string,
	renderStruct render.RenderStruct,
	opts ...interface{},
) error {
	opts = append(opts, tele.ModeHTML, tele.NoPreview)

	template, renderErr := a.TemplateManager.Render(templateName, renderStruct)

	if renderErr != nil {
		a.Logger.Error().Str("template", templateName).Err(renderErr).Msg("Error rendering template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	if editErr := c.Edit(strings.TrimSpace(template), opts...); editErr != nil {
		a.Logger.Error().Err(editErr).Msg("Error editing message")
		return editErr
	}

	return nil
}

func DefaultPrevPagePrefix(page int) string {
	return strconv.Itoa(page - 1)
}

func DefaultNextPagePrefix(page int) string {
	return strconv.Itoa(page + 1)
}

func GenerateMenuWithPagination[T any](
	chunk []T,
	textCallback func(T, int) string,
	elementPrefix string,
	elementCallback func(T) string,
	paginationPrefix string,
	page int,
	pagesTotal int,
) *tele.ReplyMarkup {
	return GenerateMenuWithPaginationAndPagePrefix(
		chunk,
		textCallback,
		elementPrefix,
		elementCallback,
		paginationPrefix,
		page,
		pagesTotal,
		DefaultPrevPagePrefix,
		DefaultNextPagePrefix,
	)
}

func GenerateMenuWithPaginationAndPagePrefix[T any](
	chunk []T,
	textCallback func(T, int) string,
	elementPrefix string,
	elementCallback func(T) string,
	paginationPrefix string,
	page int,
	pagesTotal int,
	prevPagePrefix func(int) string,
	nextPagePrefix func(int) string,
) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	rows := make([]tele.Row, 0)

	for index, element := range chunk {
		button := menu.Data(
			textCallback(element, index),
			elementPrefix,
			elementCallback(element),
		)

		rows = append(rows, menu.Row(button))
	}

	if len(chunk) > 0 {
		buttons := []tele.Btn{}
		if page >= 1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("⬅️Page %d", page),
				paginationPrefix,
				prevPagePrefix(page),
			))
		}

		if page < pagesTotal-1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("➡️Page %d", page+2),
				paginationPrefix,
				nextPagePrefix(page),
			))
		}

		if len(buttons) > 0 {
			rows = append(rows, menu.Row(buttons...))
		}
	}

	menu.Inline(rows...)
	return menu
}
