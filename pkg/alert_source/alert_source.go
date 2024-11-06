package alert_source

import "main/pkg/types"

type AlertSource interface {
	Enabled() bool
	GetAlertingRules() (types.GrafanaAlertGroups, error)
	Name() string
}
