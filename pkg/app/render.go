package app

import (
	"fmt"
	"main/pkg/utils"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleRenderPanel(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got render query")

	opts, valid := utils.ParseRenderOptions(c.Text())

	if !valid {
		return c.Reply("Usage: /render <opts> <panel name>")
	}

	panels, err := a.Grafana.GetAllPanels()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for panels: %s", err))
	}

	panel, found := panels.FindByName(opts.Query)
	if !found {
		return c.Reply("Could not find a panel. See /dashboards for dashboards list, and /dashboard <dashboard name> for its panels.")
	}

	image, err := a.Grafana.RenderPanel(panel, opts.Params)
	if err != nil {
		return a.BotReply(c, err.Error())
	}

	defer image.Close()

	fileToSend := &tele.Photo{
		File:    tele.FromReader(image),
		Caption: fmt.Sprintf("Panel: %s", a.Grafana.GetPanelLink(*panel)),
	}
	return c.Reply(fileToSend, tele.ModeHTML)
}
