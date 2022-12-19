package main

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListDatasources(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	datasources, err := a.Grafana.GetDatasources()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying datasources: %s", err))
	}

	template, err := a.TemplateManager.Render("datasources_list", RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         datasources,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering datasources_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}
