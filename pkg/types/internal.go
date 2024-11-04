package types

import (
	"main/pkg/utils/normalize"
	"strings"
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
	GroupName        string
	GroupAlertsCount int
	AlertName        string
	Alert            GrafanaAlert
	ShowAlertName    bool
}

type FiringAlertsListStruct struct {
	GrafanaAlerts         []FiringAlert
	PrometheusAlerts      []FiringAlert
	ShowGrafanaHeader     bool
	ShowPrometheusHeader  bool
	GrafanaAlertsCount    int
	PrometheusAlertsCount int
}

type SilencesListStruct struct {
	Silences      []SilenceWithAlerts
	ShowHeader    bool
	Start         int
	End           int
	SilencesCount int
}
