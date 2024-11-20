package app

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils"
	"main/pkg/utils/generic"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleRenderPanel(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got render query")

	opts, valid := utils.ParseRenderOptions(c.Text())

	if !valid {
		return a.HandleRenderPanelChooseDashboard(c, 0, false)
	}

	return a.HandleRenderPanelGeneric(c, opts)
}

func (a *App) HandleRenderChooseDashboardFromCallback(c tele.Context) error {
	callback := c.Callback()

	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("data", callback.Data).
		Msg("Got render choose dashboard query")

	page, err := strconv.Atoi(callback.Data)
	if err != nil {
		return c.Reply("Failed to parse page number from callback!")
	}

	return a.HandleRenderPanelChooseDashboard(c, page, true)
}

func (a *App) HandleRenderPanelChooseDashboard(
	c tele.Context,
	page int,
	editPrevious bool,
) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Int("page", page).
		Msg("Got render query to show dashboards")

	dashboards, err := a.Grafana.GetAllDashboards()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error fetching dashboards: %s\n", err))
	}

	if len(dashboards) == 0 {
		return c.Reply("No dashboards configured!")
	}

	dashboardsGrouped := generic.SplitArrayIntoChunks(dashboards, constants.DashboardsInOneMessage)

	chunk := []types.GrafanaDashboardInfo{}
	if page < len(dashboardsGrouped) {
		chunk = dashboardsGrouped[page]
	}

	templateData := render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.DashboardsListStruct{
			Dashboards:      chunk,
			Start:           page*constants.DashboardsInOneMessage + 1,
			End:             page*constants.DashboardsInOneMessage + len(chunk),
			DashboardsCount: len(dashboards),
		},
	}

	menu := GenerateMenu(
		chunk,
		func(elt types.GrafanaDashboardInfo, index int) string { return elt.Title },
		constants.GrafanaRenderChoosePanelPrefix,
		func(elt types.GrafanaDashboardInfo) string { return fmt.Sprintf("%s 0", elt.UID) },
		constants.GrafanaRenderChooseDashboardPrefix,
		page,
		len(dashboardsGrouped),
	)

	if editPrevious {
		return a.EditRender(c, "render_choose_dashboard", templateData, menu)
	}

	return a.ReplyRender(c, "render_choose_dashboard", templateData, menu)
}

func (a *App) HandleRenderPanelChoosePanelFromCallback(c tele.Context) error {
	callback := c.Callback()

	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("data", callback.Data).
		Msg("Got render query to choose panel")

	data := strings.SplitN(callback.Data, " ", 2)
	if len(data) != 2 {
		return c.Reply("Invalid callback provided!")
	}

	page, err := strconv.Atoi(data[1])
	if err != nil {
		return c.Reply("Failed to parse page number from callback!")
	}

	dashboard, err := a.Grafana.GetDashboard(data[0])
	if err != nil {
		return c.Reply(fmt.Sprintf("Error fetching dashboard: %s\n", err))
	}

	if len(dashboard.Dashboard.Panels) == 0 {
		return c.Reply("This dashboard has no panels!")
	}

	panelsGrouped := generic.SplitArrayIntoChunks(dashboard.Dashboard.Panels, constants.PanelsInOneMessage)

	chunk := []types.GrafanaPanel{}
	if page < len(panelsGrouped) {
		chunk = panelsGrouped[page]
	}

	templateData := render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.PanelsListStruct{
			Dashboard:   dashboard.Dashboard,
			Panels:      chunk,
			Start:       page*constants.PanelsInOneMessage + 1,
			End:         page*constants.PanelsInOneMessage + len(chunk),
			PanelsCount: len(dashboard.Dashboard.Panels),
		},
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	rows := make([]tele.Row, 0)

	for _, panel := range chunk {
		button := menu.Data(
			panel.Title,
			constants.GrafanaRenderRenderPanelPrefix,
			fmt.Sprintf("%s %d", dashboard.Dashboard.UID, panel.ID),
		)

		rows = append(rows, menu.Row(button))
	}

	if len(chunk) > 0 {
		buttons := []tele.Btn{}
		if page >= 1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("⬅️Page %d", page),
				constants.GrafanaRenderChoosePanelPrefix,
				fmt.Sprintf("%s %d", dashboard.Dashboard.UID, page-1),
			))
		}

		if page < len(panelsGrouped)-1 {
			buttons = append(buttons, menu.Data(
				fmt.Sprintf("➡️Page %d", page+2),
				constants.GrafanaRenderChoosePanelPrefix,
				fmt.Sprintf("%s %d", dashboard.Dashboard.UID, page+1),
			))
		}

		if len(buttons) > 0 {
			rows = append(rows, menu.Row(buttons...))
		}
	}

	menu.Inline(rows...)
	return a.EditRender(c, "render_choose_panel", templateData, menu)
}

func (a *App) HandleRenderPanelFromCallback(c tele.Context) error {
	callback := c.Callback()

	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("data", callback.Data).
		Msg("Got render query to render panel")

	data := strings.SplitN(callback.Data, " ", 2)
	if len(data) != 2 {
		return c.Reply("Invalid callback provided!")
	}

	dashboard, err := a.Grafana.GetDashboard(data[0])
	if err != nil {
		return c.Reply(fmt.Sprintf("Error fetching dashboard: %s\n", err))
	}

	panel, found := generic.Find(dashboard.Dashboard.Panels, func(p types.GrafanaPanel) bool {
		return strconv.Itoa(p.ID) == data[1]
	})

	if !found {
		return c.Reply("Panel not found!")
	}

	image, err := a.Grafana.RenderPanel(panel.ID, dashboard.Dashboard.UID, map[string]string{})
	if err != nil {
		return c.Reply(fmt.Sprintf("Error rendering panel: %s", err))
	}

	defer image.Close()

	fileToSend := &tele.Photo{
		File: tele.FromReader(image),
		Caption: fmt.Sprintf("Panel: %s", a.Grafana.GetPanelLink(types.PanelStruct{
			PanelID:      panel.ID,
			DashboardURL: dashboard.Meta.URL,
			Name:         panel.Title,
		})),
	}

	replyTo := c.Message().ReplyTo

	if deleteErr := a.Bot.Delete(c.Message()); deleteErr != nil {
		a.Logger.Error().Err(deleteErr).Msg("Failed to delete message")
	}

	_, sendErr := a.Bot.Reply(replyTo, fileToSend, tele.ModeHTML)
	return sendErr
}

func (a *App) HandleRenderPanelGeneric(
	c tele.Context,
	opts types.RenderOptions,
) error {
	panels, err := a.Grafana.GetAllPanels()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for panels: %s", err))
	}

	panel, found := panels.FindByName(opts.Query)
	if !found {
		return c.Reply("Could not find a panel. See /dashboards for dashboards list, and /dashboard <dashboard name> for its panels.")
	}

	image, err := a.Grafana.RenderPanel(panel.PanelID, panel.DashboardID, opts.Params)
	if err != nil {
		return c.Reply(fmt.Sprintf("Error rendering panel: %s", err))
	}

	defer image.Close()

	fileToSend := &tele.Photo{
		File:    tele.FromReader(image),
		Caption: fmt.Sprintf("Panel: %s", a.Grafana.GetPanelLink(*panel)),
	}

	return c.Reply(fileToSend, tele.ModeHTML)
}
