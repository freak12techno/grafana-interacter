package app

import (
	"errors"
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"
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
	labels, found := groups.FindLabelsByHash(alertHashToMute)
	if !found {
		return nil, errors.New("Alert was not found!")
	}

	matchers := types.QueryMatcherFromKeyValueMap(labels)
	silenceInfo, silenceErr := utils.ParseSilenceWithDuration("callback", matchers, c.Sender().FirstName, duration)
	if silenceErr != "" {
		return nil, fmt.Errorf("Error parsing silence option: %s\n", silenceErr)
	}

	return silenceInfo, nil
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
