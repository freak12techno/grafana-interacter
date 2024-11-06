package app

import (
	"fmt"
	"main/pkg/types/render"

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

	return a.ReplyRender(c, "datasources_list", render.RenderStruct{
		Grafana: a.Grafana,
		Data:    datasources,
	})
}
