package main

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func HandleRenderPanel(c tele.Context) error {
	log.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got render query")

	opts, valid := ParseRenderOptions(c.Text())

	if !valid {
		return c.Reply("Usage: /render <opts> <panel name>")
	}

	panels, err := Grafana.GetAllPanels()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying for panels: %s", err))
	}

	panel, found := FindPanelByName(panels, opts.Query)
	if !found {
		return c.Reply("Could not find a panel. See /dashboards for dashboards list, and /dashboard <dashboard name> for its panels.")
	}

	image, err := Grafana.RenderPanel(panel, opts.Params)
	if err != nil {
		return c.Reply(err)
	}

	defer image.Close()

	fileToSend := &tele.Photo{
		File:    tele.FromReader(image),
		Caption: fmt.Sprintf("Panel: %s", Grafana.GetPanelLink(*panel)),
	}
	return c.Reply(fileToSend, tele.ModeHTML)
}
