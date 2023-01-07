package render

import (
	"main/pkg/alertmanager"
	"main/pkg/grafana"
)

type RenderStruct struct {
	Grafana      *grafana.Grafana
	Alertmanager *alertmanager.Alertmanager
	Data         interface{}
}
