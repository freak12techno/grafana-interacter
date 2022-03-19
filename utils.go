package main

import (
	"regexp"
	"strings"
)

func NormalizeString(input string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(input, ""))
}

func FindDashboardByName(dashboards []GrafanaDashboardInfo, name string) (*GrafanaDashboardInfo, bool) {
	normalizedName := NormalizeString(name)

	for _, dashboard := range dashboards {
		if NormalizeString(dashboard.Title) == normalizedName {
			return &dashboard, true
		}
	}

	return nil, false
}

func FindPanelByName(panels []PanelStruct, name string) (*PanelStruct, bool) {
	normalizedName := NormalizeString(name)

	for _, panel := range panels {
		if NormalizeString(panel.Name) == normalizedName {
			return &panel, true
		}
	}

	return nil, false
}
