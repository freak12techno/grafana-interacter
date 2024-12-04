package types

import (
	"main/pkg/utils/normalize"
	"strings"
)

type GrafanaDashboardInfo struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type GrafanaDashboardsInfo []GrafanaDashboardInfo

func (i GrafanaDashboardsInfo) FindDashboardByName(name string) (*GrafanaDashboardInfo, bool) {
	normalizedName := normalize.NormalizeString(name)

	for _, dashboard := range i {
		if strings.Contains(normalize.NormalizeString(dashboard.Title), normalizedName) {
			return &dashboard, true
		}
	}

	return nil, false
}

type GrafanaDashboardResponse struct {
	Dashboard GrafanaSingleDashboard `json:"dashboard"`
	Meta      GrafanaDashboardMeta   `json:"meta"`
}

type GrafanaSingleDashboard struct {
	Title  string         `json:"title"`
	UID    string         `json:"uid"`
	Panels []GrafanaPanel `json:"panels"`
}

type GrafanaDashboardMeta struct {
	URL string `json:"url"`
}

type GrafanaPanel struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}
