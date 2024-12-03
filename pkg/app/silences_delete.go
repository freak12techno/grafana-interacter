package app

import (
	"fmt"
	"main/pkg/silence_manager"
	"main/pkg/types/render"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleDeleteSilenceViaCommand(silenceManager silence_manager.SilenceManager) func(c tele.Context) error {
	return func(c tele.Context) error {
		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Str("silence_manager", silenceManager.Name()).
			Msg("Got new delete silence query via command")

		if !silenceManager.Enabled() {
			return c.Reply(silenceManager.Name() + " is disabled.")
		}

		args := strings.SplitN(c.Text(), " ", 2)

		if len(args) <= 1 {
			return c.Reply(fmt.Sprintf("Usage: %s <silence ID or labels>", args[0]))
		}

		return a.HandleDeleteSilenceGeneric(c, silenceManager, args[1])
	}
}

func (a *App) HandleCallbackDeleteSilence(silenceManager silence_manager.SilenceManager) func(c tele.Context) error {
	return func(c tele.Context) error {
		a.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("silence_manager", silenceManager.Name()).
			Msg("Got new delete silence callback via button")

		callback := c.Callback()

		dataSplit := strings.Split(callback.Data, " ")
		if len(dataSplit) == 2 {
			_ = a.ClearKeyboard(c)
		} else {
			a.RemoveKeyboardItemByCallback(c, callback)
		}

		return a.HandleDeleteSilenceGeneric(c, silenceManager, dataSplit[0])
	}
}

func (a *App) HandleDeleteSilenceGeneric(
	c tele.Context,
	silenceManager silence_manager.SilenceManager,
	silenceID string,
) error {
	silences, silencesFetchErr := silenceManager.GetSilences()
	if silencesFetchErr != nil {
		return c.Reply(fmt.Sprintf("Error getting silence to delete: %s", silencesFetchErr))
	}

	silence, found := silences.FindByNameOrMatchers(silenceID)
	if !found {
		return c.Reply("Silence is not found by ID or matchers: " + silenceID)
	}

	if silence.Status.State == "expired" {
		return c.Reply("Silence is already deleted!")
	}

	silenceErr := silenceManager.DeleteSilence(silence.ID)
	if silenceErr != nil {
		return c.Reply(fmt.Sprintf("Error deleting silence: %s", silenceErr))
	}

	return a.ReplyRender(c, "silences_delete", render.RenderStruct{
		Grafana: a.Grafana,
		Data:    silence,
	})
}
