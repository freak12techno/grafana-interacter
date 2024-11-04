package app

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils/generic"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListSilences(silenceManager types.SilenceManager) func(c tele.Context) error {
	return func(c tele.Context) error {
		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Str("silence_manager", silenceManager.Name()).
			Msg("Got list silence query")

		return a.HandleListSilencesWithPagination(c, silenceManager, 0, false)
	}
}

func (a *App) HandleListSilencesFromCallback(silenceManager types.SilenceManager) func(c tele.Context) error {
	return func(c tele.Context) error {
		callback := c.Callback()

		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("silence_manager", silenceManager.Name()).
			Str("data", callback.Data).
			Msg("Got list silence query via callback")

		page, err := strconv.Atoi(callback.Data)
		if err != nil {
			return c.Reply("Failed to parse page number from callback!")
		}

		return a.HandleListSilencesWithPagination(c, silenceManager, page, true)
	}
}

func (a *App) HandleListSilencesWithPagination(
	c tele.Context,
	silenceManager types.SilenceManager,
	page int,
	editPrevious bool,
) error {
	if !silenceManager.Enabled() {
		return c.Reply(silenceManager.Name() + " is disabled.")
	}

	silencesWithAlerts, err := types.GetSilencesWithAlerts(silenceManager)
	if err != nil {
		return c.Reply(fmt.Sprintf("Error fetching silences: %s\n", err))
	}

	silencesGrouped := generic.SplitArrayIntoChunks(silencesWithAlerts, constants.SilencesInOneMessage)
	if len(silencesGrouped) == 0 {
		silencesGrouped = [][]types.SilenceWithAlerts{{}}
	}

	chunk := []types.SilenceWithAlerts{}
	if page < len(silencesGrouped) {
		chunk = silencesGrouped[page]
	}

	template, renderErr := a.TemplateManager.Render("silences_list", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data: types.SilencesListStruct{
			Silences:      chunk,
			ShowHeader:    true,
			Start:         page*constants.SilencesInOneMessage + 1,
			End:           page*constants.SilencesInOneMessage + len(chunk),
			SilencesCount: len(silencesWithAlerts),
		},
	})

	if renderErr != nil {
		a.Logger.Error().Err(renderErr).Msg("Error rendering silences_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	rows := make([]tele.Row, 0)

	for _, silence := range chunk {
		button := menu.Data(
			fmt.Sprintf("❌Unsilence %s", silence.Silence.ID),
			silenceManager.GetUnsilencePrefix(),
			silence.Silence.ID,
		)

		rows = append(rows, menu.Row(button))
	}

	if len(chunk) > 0 {
		buttons := []tele.Btn{}
		if page >= 1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("⬅️Page %d", page),
				silenceManager.GetPaginatedSilencesListPrefix(),
				strconv.Itoa(page-1),
			))
		}

		if page < len(silencesGrouped)-1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("➡️Page %d", page+2),
				silenceManager.GetPaginatedSilencesListPrefix(),
				strconv.Itoa(page+1),
			))
		}

		if len(buttons) > 0 {
			rows = append(rows, menu.Row(buttons...))
		}
	}

	menu.Inline(rows...)

	if editPrevious {
		if editErr := c.Edit(template, menu, tele.ModeHTML, tele.NoPreview); editErr != nil {
			a.Logger.Error().Err(editErr).Msg("Error deleting previous message")
			return editErr
		}

		return nil
	}

	if sendErr := a.BotReply(c, template, menu); sendErr != nil {
		a.Logger.Error().Err(sendErr).Msg("Error sending message")
		return sendErr
	}

	return nil
}
