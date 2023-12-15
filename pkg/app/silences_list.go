package app

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"
	"main/pkg/utils/generic"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListSilences(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got list silence query")

	silences, err := a.Grafana.GetSilences()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error listing silence: %s", err))
	}

	silences = generic.Filter(silences, func(s types.Silence) bool {
		return s.EndsAt.After(time.Now())
	})

	silencesRender := make([]render.SilenceRender, len(silences))
	for index, silence := range silences {
		silencesRender[index] = render.SilenceRender{
			Silence:       silence,
			AlertsPresent: false,
			Alerts:        make([]types.AlertmanagerAlert, 0),
		}
	}

	template, err := a.TemplateManager.Render("silences_list", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         silencesRender,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering silences_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}

func (a *App) HandleAlertmanagerListSilences(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got Alertmanager list silence query")

	if !a.Alertmanager.Enabled() {
		return c.Reply("Alertmanager is disabled.")
	}

	silences, err := a.Alertmanager.GetSilences()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error listing silence: %s", err))
	}

	silences = generic.Filter(silences, func(s types.Silence) bool {
		return s.EndsAt.After(time.Now())
	})

	silencesRender := make([]render.SilenceRender, len(silences))

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errs []error

	for index, silence := range silences {
		wg.Add(1)
		go func(index int, silence types.Silence) {
			defer wg.Done()

			alerts, alertsErr := a.Alertmanager.GetSilenceMatchingAlerts(silence)
			if alertsErr != nil {
				mutex.Lock()
				errs = append(errs, alertsErr)
				mutex.Unlock()
				return
			}

			mutex.Lock()
			silencesRender[index] = render.SilenceRender{
				Silence:       silence,
				AlertsPresent: true,
				Alerts:        alerts,
			}
			mutex.Unlock()
		}(index, silence)
	}

	wg.Wait()

	if len(errs) > 0 {
		return c.Reply(fmt.Sprintf("Error getting alerts for silence on %d silences", len(errs)))
	}

	template, err := a.TemplateManager.Render("silences_list", render.RenderStruct{
		Grafana:      a.Grafana,
		Alertmanager: a.Alertmanager,
		Data:         silencesRender,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error rendering silences_list template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	return a.BotReply(c, template)
}
