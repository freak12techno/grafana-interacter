package render

import (
	"main/pkg/alertmanager"
	"main/pkg/grafana"
	"main/pkg/types"
)

type RenderStruct struct {
	Grafana      *grafana.Grafana
	Alertmanager *alertmanager.Alertmanager
	Data         interface{}
}

type SilenceRender struct {
	Silence       types.Silence
	AlertsPresent bool
	Alerts        []types.AlertmanagerAlert
}
