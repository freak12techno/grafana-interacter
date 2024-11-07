package app

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils/generic"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleShowDashboard(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboard query")

	args := strings.SplitN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /render

	if len(args) != 1 {
		return c.Reply("Usage: /dashboard <dashboard>")
	}

	dashboards, err := a.Grafana.GetAllDashboards()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for dashboards: %s", err))
	}

	dashboard, found := dashboards.FindDashboardByName(args[0])
	if !found {
		return c.Reply("Could not find dashboard. See /dashboards for dashboards list.")
	}

	dashboardEnriched, err := a.Grafana.GetDashboard(dashboard.UID)
	if err != nil {
		return c.Reply(fmt.Sprintf("Could not get dashboard: %s", err))
	}

	return a.ReplyRender(c, "dashboards_show", render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.DashboardStruct{
			Dashboard: *dashboard,
			Panels: generic.Map(dashboardEnriched.Dashboard.Panels, func(p types.GrafanaPanel) types.PanelStruct {
				return types.PanelStruct{
					DashboardURL: dashboard.URL,
					PanelID:      p.ID,
					Name:         p.Title,
				}
			}),
		},
	})
}
