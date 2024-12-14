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

type AlertsListForAlertSourceStruct struct {
	AlertSourceName string
	AlertGroups     []GrafanaAlertGroup
}

type AlertsListStruct struct {
	AlertSources []AlertsListForAlertSourceStruct
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
	RenderTime      time.Time
}

func (f FiringAlertsListStruct) GetAlertFiringFor(alert FiringAlert) time.Duration {
	return f.RenderTime.Sub(alert.Alert.ActiveAt)
}

type SilencesListStruct struct {
	Silences      []SilenceWithAlerts
	Start         int
	End           int
	SilencesCount int
}

type DashboardsListStruct struct {
	Dashboards      []GrafanaDashboardInfo
	Start           int
	End             int
	DashboardsCount int
}

type PanelsListStruct struct {
	Dashboard   GrafanaSingleDashboard
	Panels      []GrafanaPanel
	Start       int
	End         int
	PanelsCount int
}

type SingleAlertStruct struct {
	Alert      *GrafanaAlertRule
	RenderTime time.Time
}

func (s SingleAlertStruct) GetAlertFiringFor(alert GrafanaAlert) time.Duration {
	return s.RenderTime.Sub(alert.ActiveAt)
}

type SilencePrepareStruct struct {
	Matchers    QueryMatchers
	AlertsCount int
}
