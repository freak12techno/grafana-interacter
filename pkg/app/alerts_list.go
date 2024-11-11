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

	alertSourcesGroups := []types.AlertsListForAlertSourceStruct{}

	for _, alertSource := range a.AlertSourcesWithSilenceManager {
		if !alertSource.AlertSource.Enabled() {
			continue
		}

		alertGroups, err := alertSource.AlertSource.GetAlertingRules()
		if err != nil {
			return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
		}

		alertSourcesGroups = append(alertSourcesGroups, types.AlertsListForAlertSourceStruct{
			AlertSourceName: alertSource.AlertSource.Name(),
			AlertGroups:     alertGroups,
		})
	}

	return a.ReplyRender(c, "alerts_list", render.RenderStruct{
		Grafana: a.Grafana,
		Data:    alertSourcesGroups,
	})
}
