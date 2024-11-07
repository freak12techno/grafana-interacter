package render

import (
	"main/pkg/clients"
)

type RenderStruct struct {
	Grafana *clients.Grafana
	Data    interface{}
}
