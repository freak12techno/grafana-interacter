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

func (a *App) HandleChooseSilenceManagerForListSilences(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got choosing a datasource for list silences query")

	silenceManagers := generic.Filter(a.AlertSourcesWithSilenceManager, func(a AlertSourceWithSilenceManager) bool {
		return a.SilenceManager.Enabled()
	})

	if len(silenceManagers) == 0 {
		return a.BotReply(c, "No silence managers configured!")
	}

	if len(silenceManagers) == 1 {
		return a.HandleListSilencesWithPagination(
			c,
			silenceManagers[0].SilenceManager,
			0,
			false,
		)
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	rows := make([]tele.Row, 0)
	index := 0

	for _, source := range silenceManagers {
		button := menu.Data(
			source.SilenceManager.Name(),
			source.SilenceManager.Prefixes().PaginatedSilencesList,
			"0", // page
		)

		rows = append(rows, menu.Row(button))
		index += 1
	}

	menu.Inline(rows...)

	return a.BotReply(c, "Choose a silence manager to get silences from:", menu)
}

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

	silencesWithAlerts, totalCount, err := silence_manager.GetSilencesWithAlerts(
		silenceManager,
		page,
		constants.SilencesInOneMessage,
	)
	if err != nil {
		return c.Reply(fmt.Sprintf("Error fetching silences: %s\n", err))
	}

	templateData := render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.SilencesListStruct{
			Silences:      silencesWithAlerts,
			Start:         page*constants.SilencesInOneMessage + 1,
			End:           page*constants.SilencesInOneMessage + len(silencesWithAlerts),
			SilencesCount: totalCount,
		},
	}

	prefixes := silenceManager.Prefixes()

	menu := GenerateMenuWithPagination(
		silencesWithAlerts,
		func(elt types.SilenceWithAlerts, index int) string {
			return fmt.Sprintf("❌Unsilence %s", elt.Silence.ID)
		},
		prefixes.Unsilence,
		func(elt types.SilenceWithAlerts) string { return elt.Silence.ID },
		prefixes.PaginatedSilencesList,
		page,
		totalCount,
	)

	if editPrevious {
		return a.EditRender(c, "silences_list", templateData, menu)
	}

	return a.ReplyRender(c, "silences_list", templateData, menu)
}
