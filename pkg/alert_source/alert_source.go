package alert_source

import "main/pkg/types"

type Prefixes struct {
	PaginatedFiringAlerts string
}

type AlertSource interface {
	Enabled() bool
	GetAlertingRules() (types.GrafanaAlertGroups, error)
	Name() string
	Prefixes() Prefixes
}
