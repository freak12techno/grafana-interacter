package app

import (
	"main/pkg/types/render"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleHelp(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got help query")

	return a.ReplyRender(c, "help", render.RenderStruct{
		Grafana: a.Grafana,
		Data:    a.Version,
	})
}
