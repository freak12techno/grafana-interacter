package app

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListAlerts(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got alerts query")

	// TODO: fix
	grafanaGroups, err := a.AlertSourcesWithSilenceManager[0].AlertSource.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	prometheusGroups, err := a.AlertSourcesWithSilenceManager[1].AlertSource.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	return a.ReplyRender(c, "alerts_list", render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.AlertsListStruct{
			GrafanaGroups:    grafanaGroups,
			PrometheusGroups: prometheusGroups,
		},
	})
}
