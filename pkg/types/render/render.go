package render

import (
	"main/pkg/clients"
	"main/pkg/silence_manager"
)

type RenderStruct struct {
	Grafana      *clients.Grafana
	Alertmanager *silence_manager.Alertmanager
	Data         interface{}
}
