package app

import (
	"fmt"
	"main/pkg/alert_source"
	"main/pkg/constants"
	"main/pkg/silence_manager"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils/generic"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleChooseAlertSourceForListFiringAlerts(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got choosing a datasource for firing alerts query")

	alertSources := generic.Filter(a.AlertSourcesWithSilenceManager, func(a AlertSourceWithSilenceManager) bool {
		return a.AlertSource.Enabled()
	})

	if len(alertSources) == 0 {
		return a.BotReply(c, "No alert sources configured!")
	}

	if len(alertSources) == 1 {
		return a.HandleListFiringAlertsWithPagination(
			c,
			alertSources[0].AlertSource,
			alertSources[0].SilenceManager,
			0,
			false,
		)
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	rows := make([]tele.Row, 0)
	index := 0

	for _, source := range alertSources {
		if !source.AlertSource.Enabled() {
			continue
		}

		button := menu.Data(
			source.AlertSource.Name(),
			source.AlertSource.Prefixes().PaginatedFiringAlerts,
			"0", // page
		)

		rows = append(rows, menu.Row(button))
		index += 1
	}

	menu.Inline(rows...)

	return a.BotReply(c, "Choose an alert source to get alerts from:", menu)
}

func (a *App) HandleListFiringAlertsFromCallback(
	alertSource alert_source.AlertSource,
	silenceManager silence_manager.SilenceManager,
) func(c tele.Context) error {
	return func(c tele.Context) error {
		callback := c.Callback()

		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("alert_source", alertSource.Name()).
			Str("data", callback.Data).
			Msg("Got list firing alerts query via callback")

		page, err := strconv.Atoi(callback.Data)
		if err != nil {
			return c.Reply("Failed to parse page number from callback!")
		}

		return a.HandleListFiringAlertsWithPagination(c, alertSource, silenceManager, page, true)
	}
}

func (a *App) HandleListFiringAlertsWithPagination(
	c tele.Context,
	alertSource alert_source.AlertSource,
	silenceManager silence_manager.SilenceManager,
	page int,
	editPrevious bool,
) error {
	if !alertSource.Enabled() {
		return c.Reply(alertSource.Name() + " is disabled.")
	}

	alerts, err := alertSource.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error fetching alerts: %s!\n", err))
	}

	alerts = alerts.FilterFiringOrPendingAlertGroups()

	firingAlerts := make([]types.FiringAlert, 0)

	for _, alertGroup := range alerts {
		for _, alertRule := range alertGroup.Rules {
			for _, alert := range alertRule.Alerts {
				firingAlerts = append(firingAlerts, types.FiringAlert{
					GroupName:     alertGroup.Name,
					Alert:         alert,
					AlertRuleName: alertRule.Name,
				})
			}
		}
	}

	alertsGrouped := generic.SplitArrayIntoChunks(firingAlerts, constants.AlertsInOneMessage)
	if len(alertsGrouped) == 0 {
		alertsGrouped = [][]types.FiringAlert{{}}
	}

	chunk := []types.FiringAlert{}
	if page < len(alertsGrouped) {
		chunk = alertsGrouped[page]
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	rows := make([]tele.Row, 0)
	index := 0

	for _, alert := range chunk {
		button := menu.Data(
			fmt.Sprintf("🔇Silence alert #%d", index+1),
			silenceManager.Prefixes().PrepareSilence,
			alert.Alert.GetCallbackHash(),
		)

		rows = append(rows, menu.Row(button))
		index += 1
	}

	if len(chunk) > 0 {
		buttons := []tele.Btn{}
		if page >= 1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("⬅️Page %d", page),
				alertSource.Prefixes().PaginatedFiringAlerts,
				strconv.Itoa(page-1),
			))
		}

		if page < len(alertsGrouped)-1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("➡️Page %d", page+2),
				alertSource.Prefixes().PaginatedFiringAlerts,
				strconv.Itoa(page+1),
			))
		}

		if len(buttons) > 0 {
			rows = append(rows, menu.Row(buttons...))
		}
	}

	menu.Inline(rows...)

	templateData := render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.FiringAlertsListStruct{
			AlertSourceName: alertSource.Name(),
			Alerts:          chunk,
			AlertsCount:     len(firingAlerts),
			Start:           page*constants.AlertsInOneMessage + 1,
			End:             page*constants.AlertsInOneMessage + len(chunk),
		},
	}

	if editPrevious {
		return a.EditRender(c, "alerts_firing", templateData, menu)
	}

	return a.ReplyRender(c, "alerts_firing", templateData, menu)
}
