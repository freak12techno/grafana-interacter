package types

import (
	"main/pkg/utils/normalize"
	"strings"
	"time"
)

type DashboardStruct struct {
	Dashboard GrafanaDashboardInfo
	Panels    []PanelStruct
}

type PanelStruct struct {
	Name          string
	DashboardName string
	DashboardID   string
	DashboardURL  string
	PanelID       int
}

type PanelsStruct []PanelStruct

func (s PanelsStruct) FindByName(name string) (*PanelStruct, bool) {
	normalizedName := normalize.NormalizeString(name)

	for _, panel := range s {
		panelNameWithDashboardName := normalize.NormalizeString(panel.DashboardName + panel.Name)

		if strings.Contains(panelNameWithDashboardName, normalizedName) {
			return &panel, true
		}
	}

	return nil, false
}

type AlertsListStruct struct {
	GrafanaGroups    []GrafanaAlertGroup
	PrometheusGroups []GrafanaAlertGroup
}

type FiringAlert struct {
	GroupName     string
	AlertRuleName string
	Alert         GrafanaAlert
}

type FiringAlertsListStruct struct {
	AlertSourceName string
	Alerts          []FiringAlert
	AlertsCount     int
	Start           int
	End             int
}

type SilencesListStruct struct {
	Silences      []SilenceWithAlerts
	ShowHeader    bool
	Start         int
	End           int
	SilencesCount int
}

type SingleAlertStruct struct {
	Alert      *GrafanaAlertRule
	RenderTime time.Time
}

func (s SingleAlertStruct) GetAlertFiringFor(alert GrafanaAlert) time.Duration {
	return s.RenderTime.Sub(alert.ActiveAt)
}
