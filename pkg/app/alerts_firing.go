package app

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/types/render"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleListFiringAlerts(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got firing alerts query")

	grafanaGroups, err := a.Grafana.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	prometheusGroups, err := a.Prometheus.GetAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	grafanaGroups = grafanaGroups.FilterFiringOrPendingAlertGroups()
	prometheusGroups = prometheusGroups.FilterFiringOrPendingAlertGroups()

	batches := []types.FiringAlertsListStruct{}
	batchToAdd := types.FiringAlertsListStruct{
		GrafanaAlerts:         make([]types.FiringAlert, 0),
		PrometheusAlerts:      make([]types.FiringAlert, 0),
		GrafanaAlertsCount:    len(grafanaGroups),
		PrometheusAlertsCount: len(prometheusGroups),
		ShowGrafanaHeader:     true,
	}
	batchIndex := 0

	for _, grafanaGroup := range grafanaGroups {
		for ruleIndex, grafanaRule := range grafanaGroup.Rules {
			for _, grafanaAlert := range grafanaRule.Alerts {
				batchToAdd.GrafanaAlerts = append(batchToAdd.GrafanaAlerts, types.FiringAlert{
					GroupName:        grafanaGroup.Name,
					GroupAlertsCount: len(grafanaGroup.Rules),
					AlertName:        grafanaRule.Name,
					Alert:            grafanaAlert,
					ShowAlertName:    ruleIndex == 0,
				})
				batchIndex++

				if len(batchToAdd.GrafanaAlerts) >= constants.AlertsInOneMessage {
					batches = append(batches, batchToAdd)
					batchToAdd = types.FiringAlertsListStruct{
						GrafanaAlerts:         make([]types.FiringAlert, 0),
						PrometheusAlerts:      make([]types.FiringAlert, 0),
						GrafanaAlertsCount:    len(grafanaGroups),
						PrometheusAlertsCount: len(prometheusGroups),
					}
					batchIndex = 0
				}
			}
		}
	}

	batchToAdd.ShowPrometheusHeader = true

	for _, prometheusGroup := range prometheusGroups {
		for _, prometheusRule := range prometheusGroup.Rules {
			for alertIndex, prometheusAlert := range prometheusRule.Alerts {
				batchToAdd.PrometheusAlerts = append(batchToAdd.PrometheusAlerts, types.FiringAlert{
					GroupName:        prometheusGroup.Name,
					GroupAlertsCount: len(prometheusGroup.Rules),
					AlertName:        prometheusRule.Name,
					Alert:            prometheusAlert,
					ShowAlertName:    alertIndex == 0,
				})
				batchIndex++

				if len(batchToAdd.PrometheusAlerts) >= constants.AlertsInOneMessage {
					batches = append(batches, batchToAdd)
					batchToAdd = types.FiringAlertsListStruct{
						GrafanaAlerts:         make([]types.FiringAlert, 0),
						PrometheusAlerts:      make([]types.FiringAlert, 0),
						GrafanaAlertsCount:    len(grafanaGroups),
						PrometheusAlertsCount: len(prometheusGroups),
					}
					batchIndex = 0
				}
			}
		}
	}

	if len(batches) == 0 {
		batches = append(batches, batchToAdd)
	}

	for _, batch := range batches {
		template, renderErr := a.TemplateManager.Render("alerts_firing", render.RenderStruct{
			Grafana:      a.Grafana,
			Alertmanager: a.Alertmanager,
			Data:         batch,
		})
		if renderErr != nil {
			a.Logger.Error().Err(renderErr).Msg("Error rendering alerts_firing template")
			return c.Reply(fmt.Sprintf("Error rendering template: %s", renderErr))
		}

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}

		rows := make([]tele.Row, 0)

		index := 0

		for _, alert := range batch.GrafanaAlerts {
			button := menu.Data(
				fmt.Sprintf("🔇Silence alert #%d", index+1),
				constants.AlertmanagerPrepareSilencePrefix,
				alert.Alert.GetCallbackHash(),
			)

			rows = append(rows, menu.Row(button))
			index += 1
		}

		for _, alert := range batch.PrometheusAlerts {
			button := menu.Data(
				fmt.Sprintf("🔇Silence alert #%d", index+1),
				constants.AlertmanagerPrepareSilencePrefix,
				alert.Alert.GetCallbackHash(),
			)

			rows = append(rows, menu.Row(button))
			index += 1
		}

		menu.Inline(rows...)

		if sendErr := a.BotReply(c, template, menu); sendErr != nil {
			return err
		}
	}

	return nil
}
