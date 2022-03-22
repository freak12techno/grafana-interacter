package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type GrafanaStruct struct {
	URL    string
	Auth   *AuthStruct
	Logger zerolog.Logger
}

func InitGrafana(url string, auth *AuthStruct, logger *zerolog.Logger) *GrafanaStruct {
	return &GrafanaStruct{
		URL:    url,
		Auth:   auth,
		Logger: logger.With().Str("component", "grafanaStruct").Logger(),
	}
}

func (g *GrafanaStruct) UseAuth() bool {
	return g.Auth != nil && g.Auth.User != "" && g.Auth.Password != ""
}

func (g *GrafanaStruct) RenderPanel(panel *PanelStruct, qs map[string]string) (io.ReadCloser, error) {
	baseParams := map[string]string{
		"orgId":   "1",
		"from":    "now",
		"to":      "now-30m",
		"panelId": fmt.Sprintf("%d", panel.PanelID),
		"width":   "1000",
		"height":  "500",
		"tz":      "Europe/Moscow",
	}

	url := fmt.Sprintf(
		"%s/render/d-solo/%s/dashboard?%s",
		g.URL,
		panel.DashboardID,
		SerializeQueryString(MergeMaps(baseParams, qs)),
	)

	return g.Query(url)
}

func (g *GrafanaStruct) GetAllDashboards() ([]GrafanaDashboardInfo, error) {
	url := fmt.Sprintf("%s/api/search?type=dash-db", g.URL)
	dashboards := []GrafanaDashboardInfo{}
	err := g.QueryAndDecode(url, &dashboards)
	return dashboards, err
}

func (g *GrafanaStruct) GetDashboard(dashboardUID string) (*GrafanaDashboardResponse, error) {
	url := fmt.Sprintf("%s/api/dashboards/uid/%s", g.URL, dashboardUID)
	dashboards := &GrafanaDashboardResponse{}
	err := g.QueryAndDecode(url, &dashboards)
	return dashboards, err
}

func (g *GrafanaStruct) GetAllPanels() ([]PanelStruct, error) {
	dashboards, err := g.GetAllDashboards()
	if err != nil {
		return nil, err
	}

	dashboardsEnriched := make([]GrafanaDashboardResponse, len(dashboards))
	group, _ := errgroup.WithContext(context.Background())

	for i, d := range dashboards {
		index := i
		dashboard := d

		group.Go(func() error {
			enrichedDashboard, err := g.GetDashboard(dashboard.UID)
			if err == nil {
				dashboardsEnriched[index] = *enrichedDashboard
			}

			return err
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	panelsCount := 0
	for _, d := range dashboardsEnriched {
		panelsCount += len(d.Dashboard.Panels)
	}

	panels := make([]PanelStruct, panelsCount)
	counter := 0

	for _, d := range dashboardsEnriched {
		for _, p := range d.Dashboard.Panels {
			panels[counter] = PanelStruct{
				Name:          p.Title,
				DashboardName: d.Dashboard.Title,
				DashboardID:   d.Dashboard.UID,
				DashboardURL:  d.Meta.URL,
				PanelID:       p.ID,
			}

			counter++
		}
	}

	return panels, nil
}

func (g *GrafanaStruct) GetDashboardLink(dashboard GrafanaDashboardInfo) string {
	return fmt.Sprintf("<a href='%s%s'>%s</a>", g.URL, dashboard.URL, dashboard.Title)
}

func (g *GrafanaStruct) GetPanelLink(panel PanelStruct) string {
	return fmt.Sprintf(
		"<a href='%s%s?viewPanel=%d'>%s</a>",
		Config.GrafanaURL,
		panel.DashboardURL,
		panel.PanelID,
		panel.Name,
	)
}

func (g *GrafanaStruct) GetDatasourceLink(ds GrafanaDatasource) string {
	return fmt.Sprintf("<a href='%s/datasources/edit/%s'>%s</a>", g.URL, ds.UID, ds.Name)
}

func (g *GrafanaStruct) GetDatasources() ([]GrafanaDatasource, error) {
	datasources := []GrafanaDatasource{}
	url := fmt.Sprintf("%s/api/datasources", g.URL)
	err := g.QueryAndDecode(url, &datasources)
	return datasources, err
}

func (g *GrafanaStruct) GetGrafanaAlertingRules() ([]GrafanaAlertGroup, error) {
	rules := GrafanaAlertRulesResponse{}
	url := fmt.Sprintf("%s/api/prometheus/grafana/api/v1/rules", g.URL)
	err := g.QueryAndDecode(url, &rules)
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}

func (g *GrafanaStruct) GetDatasourceAlertingRules(datasourceID int) ([]GrafanaAlertGroup, error) {
	rules := GrafanaAlertRulesResponse{}
	url := fmt.Sprintf("%s/api/prometheus/%d/api/v1/rules", g.URL, datasourceID)
	err := g.QueryAndDecode(url, &rules)
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}

func (g *GrafanaStruct) GetPrometheusAlertingRules() ([]GrafanaAlertGroup, error) {
	datasources, err := g.GetDatasources()
	if err != nil {
		return nil, err
	}

	groups := []GrafanaAlertGroup{}
	for _, ds := range datasources {
		if ds.Type == "prometheus" {
			resp, err := g.GetDatasourceAlertingRules(ds.ID)
			if err != nil {
				return nil, err
			}

			groups = append(groups, resp...)
		}
	}

	return groups, err
}

func (g *GrafanaStruct) Query(url string) (io.ReadCloser, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	g.Logger.Trace().Str("url", url).Msg("Doing a Grafana API query")

	if g.UseAuth() {
		req.SetBasicAuth(g.Auth.User, g.Auth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Could not fetch request. Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (g *GrafanaStruct) QueryAndDecode(url string, output interface{}) error {
	body, err := g.Query(url)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}
