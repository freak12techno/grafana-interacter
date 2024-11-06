package app

import (
	"fmt"
	"main/pkg/types/render"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListDashboards(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got dashboards query")

	dashboards, err := a.Grafana.GetAllDashboards()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying dashboards: %s", err))
	}

	template, err := a.TemplateManager.Render("dashboards_list", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         dashboards,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering dashboards_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}
