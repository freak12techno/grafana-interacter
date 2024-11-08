package app

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/silence_manager"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils/generic"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListSilences(silenceManager silence_manager.SilenceManager) func(c tele.Context) error {
	return func(c tele.Context) error {
		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Str("silence_manager", silenceManager.Name()).
			Msg("Got list silence query")

		return a.HandleListSilencesWithPagination(c, silenceManager, 0, false)
	}
}

func (a *App) HandleListSilencesFromCallback(silenceManager silence_manager.SilenceManager) func(c tele.Context) error {
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
	silenceManager silence_manager.SilenceManager,
	page int,
	editPrevious bool,
) error {
	if !silenceManager.Enabled() {
		return c.Reply(silenceManager.Name() + " is disabled.")
	}

	silencesWithAlerts, err := silence_manager.GetSilencesWithAlerts(silenceManager)
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

	templateData := render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.SilencesListStruct{
			Silences:      chunk,
			ShowHeader:    true,
			Start:         page*constants.SilencesInOneMessage + 1,
			End:           page*constants.SilencesInOneMessage + len(chunk),
			SilencesCount: len(silencesWithAlerts),
		},
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
		return a.EditRenderWithMarkup(c, "silences_list", templateData, menu)
	}

	return a.ReplyRenderWithMarkup(c, "silences_list", templateData, menu)
}
