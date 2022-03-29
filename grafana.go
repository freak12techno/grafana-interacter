package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type GrafanaStruct struct {
	Config *GrafanaConfig
	Logger zerolog.Logger
}

func InitGrafana(config *GrafanaConfig, logger *zerolog.Logger) *GrafanaStruct {
	return &GrafanaStruct{
		Config: config,
		Logger: logger.With().Str("component", "grafana").Logger(),
	}
}

func (g *GrafanaStruct) UseAuth() bool {
	return g.Config != nil && g.Config.User != "" && g.Config.Password != ""
}

func (g *GrafanaStruct) RenderPanel(panel *PanelStruct, qs map[string]string) (io.ReadCloser, error) {
	baseParams := map[string]string{
		"orgId":   "1",
		"from":    "now",
		"to":      "now-30m",
		"panelId": fmt.Sprintf("%d", panel.PanelID),
		"width":   "1000",
		"height":  "500",
		"tz":      g.Config.Timezone,
	}

	url := g.RelativeLink(fmt.Sprintf(
		"/render/d-solo/%s/dashboard?%s",
		panel.DashboardID,
		SerializeQueryString(MergeMaps(baseParams, qs)),
	))

	return g.Query(url)
}

func (g *GrafanaStruct) GetAllDashboards() ([]GrafanaDashboardInfo, error) {
	url := g.RelativeLink("/api/search?type=dash-db")
	dashboards := []GrafanaDashboardInfo{}
	err := g.QueryAndDecode(url, &dashboards)
	return dashboards, err
}

func (g *GrafanaStruct) GetDashboard(dashboardUID string) (*GrafanaDashboardResponse, error) {
	url := g.RelativeLink(fmt.Sprintf("/api/dashboards/uid/%s", dashboardUID))
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

func (g *GrafanaStruct) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
}

func (g *GrafanaStruct) GetDashboardLink(dashboard GrafanaDashboardInfo) string {
	return fmt.Sprintf("<a href='%s%s'>%s</a>", g.Config.URL, dashboard.URL, dashboard.Title)
}

func (g *GrafanaStruct) GetPanelLink(panel PanelStruct) string {
	return fmt.Sprintf(
		"<a href='%s?viewPanel=%d'>%s</a>",
		g.RelativeLink(panel.DashboardURL),
		panel.PanelID,
		panel.Name,
	)
}

func (g *GrafanaStruct) GetDatasourceLink(ds GrafanaDatasource) string {
	return fmt.Sprintf("<a href='%s/datasources/edit/%s'>%s</a>", g.Config.URL, ds.UID, ds.Name)
}

func (g *GrafanaStruct) GetDatasources() ([]GrafanaDatasource, error) {
	datasources := []GrafanaDatasource{}
	url := g.RelativeLink("/api/datasources")
	err := g.QueryAndDecode(url, &datasources)
	return datasources, err
}

func (g *GrafanaStruct) GetGrafanaAlertingRules() ([]GrafanaAlertGroup, error) {
	rules := GrafanaAlertRulesResponse{}
	url := g.RelativeLink("/api/prometheus/grafana/api/v1/rules")
	err := g.QueryAndDecode(url, &rules)
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}

func (g *GrafanaStruct) GetDatasourceAlertingRules(datasourceID int) ([]GrafanaAlertGroup, error) {
	rules := GrafanaAlertRulesResponse{}
	url := g.RelativeLink(fmt.Sprintf("/api/prometheus/%d/api/v1/rules", datasourceID))
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

func (g *GrafanaStruct) GetAllAlertingRules() ([]GrafanaAlertGroup, error) {
	grafanaRules, err := g.GetGrafanaAlertingRules()
	if err != nil {
		return nil, err
	}

	prometheusRules, err := g.GetPrometheusAlertingRules()
	if err != nil {
		return nil, err
	}

	return append(grafanaRules, prometheusRules...), nil
}

func (g *GrafanaStruct) CreateSilence(silence Silence) (Silence, error) {
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	res := Silence{}
	err := g.QueryAndDecodePost(url, silence, res)
	return res, err
}

func (g *GrafanaStruct) GetSilences() ([]Silence, error) {
	silences := []Silence{}
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	err := g.QueryAndDecode(url, &silences)
	return silences, err
}

func (g *GrafanaStruct) Query(url string) (io.ReadCloser, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	g.Logger.Trace().Str("url", url).Msg("Doing a Grafana API query")

	if g.UseAuth() {
		req.SetBasicAuth(g.Config.User, g.Config.Password)
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

func (g *GrafanaStruct) QueryPost(url string, body interface{}) (io.ReadCloser, error) {
	client := &http.Client{}

	buffer := new(bytes.Buffer)

	if err := json.NewEncoder(buffer).Encode(body); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	g.Logger.Trace().Str("url", url).Msg("Doing a Grafana API query")

	if g.UseAuth() {
		req.SetBasicAuth(g.Config.User, g.Config.Password)
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

func (g *GrafanaStruct) QueryAndDecodePost(url string, postBody interface{}, output interface{}) error {
	body, err := g.QueryPost(url, postBody)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}
