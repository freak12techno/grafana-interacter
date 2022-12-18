package main

import (
	"fmt"
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

	dashboard, found := FindDashboardByName(dashboards, args[0])
	if !found {
		return c.Reply("Could not find dashboard. See /dashboards for dashboards list.")
	}

	dashboardEnriched, err := a.Grafana.GetDashboard(dashboard.UID)
	if err != nil {
		return c.Reply(fmt.Sprintf("Could not get dashboard: %s", err))
	}

	template, err := a.TemplateManager.Render("dashboards_show", RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data: DashboardStruct{
			Dashboard: *dashboard,
			Panels: Map(dashboardEnriched.Dashboard.Panels, func(p GrafanaPanel) PanelStruct {
				return PanelStruct{
					DashboardURL: dashboard.URL,
					PanelID:      p.ID,
					Name:         p.Title,
				}
			}),
		},
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering dashboards_show template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}
