package types

type AlertSource interface {
	Enabled() bool
	GetAlertingRules() (GrafanaAlertGroups, error)
	AlertSourceName() string
}
