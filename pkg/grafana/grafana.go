package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"main/pkg/config"
	"main/pkg/types"
	"main/pkg/utils"
	"net/http"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Grafana struct {
	Config config.GrafanaConfig
	Logger zerolog.Logger
}

func InitGrafana(config config.GrafanaConfig, logger *zerolog.Logger) *Grafana {
	return &Grafana{
		Config: config,
		Logger: logger.With().Str("component", "grafana").Logger(),
	}
}

func (g *Grafana) UseAuth() bool {
	return g.Config.User != "" && g.Config.Password != ""
}

func (g *Grafana) RenderPanel(panel *types.PanelStruct, qs map[string]string) (io.ReadCloser, error) {
	params := utils.MergeMaps(g.Config.RenderOptions, qs)
	params["panelId"] = fmt.Sprintf("%d", panel.PanelID)

	url := g.RelativeLink(fmt.Sprintf(
		"/render/d-solo/%s/dashboard?%s",
		panel.DashboardID,
		utils.SerializeQueryString(params),
	))

	return g.Query(url)
}

func (g *Grafana) GetAllDashboards() ([]types.GrafanaDashboardInfo, error) {
	url := g.RelativeLink("/api/search?type=dash-db")
	dashboards := []types.GrafanaDashboardInfo{}
	err := g.QueryAndDecode(url, &dashboards)
	return dashboards, err
}

func (g *Grafana) GetDashboard(dashboardUID string) (*types.GrafanaDashboardResponse, error) {
	url := g.RelativeLink(fmt.Sprintf("/api/dashboards/uid/%s", dashboardUID))
	dashboards := &types.GrafanaDashboardResponse{}
	err := g.QueryAndDecode(url, &dashboards)
	return dashboards, err
}

func (g *Grafana) GetAllPanels() ([]types.PanelStruct, error) {
	dashboards, err := g.GetAllDashboards()
	if err != nil {
		return nil, err
	}

	dashboardsEnriched := make([]types.GrafanaDashboardResponse, len(dashboards))
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

	panels := make([]types.PanelStruct, panelsCount)
	counter := 0

	for _, d := range dashboardsEnriched {
		for _, p := range d.Dashboard.Panels {
			panels[counter] = types.PanelStruct{
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

func (g *Grafana) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
}

func (g *Grafana) GetDashboardLink(dashboard types.GrafanaDashboardInfo) template.HTML {
	return template.HTML(fmt.Sprintf("<a href='%s%s'>%s</a>", g.Config.URL, dashboard.URL, dashboard.Title))
}

func (g *Grafana) GetPanelLink(panel types.PanelStruct) template.HTML {
	return template.HTML(fmt.Sprintf(
		"<a href='%s?viewPanel=%d'>%s</a>",
		g.RelativeLink(panel.DashboardURL),
		panel.PanelID,
		panel.Name,
	))
}

func (g *Grafana) GetDatasourceLink(ds types.GrafanaDatasource) template.HTML {
	return template.HTML(fmt.Sprintf("<a href='%s/datasources/edit/%s'>%s</a>", g.Config.URL, ds.UID, ds.Name))
}

func (g *Grafana) GetDatasources() ([]types.GrafanaDatasource, error) {
	datasources := []types.GrafanaDatasource{}
	url := g.RelativeLink("/api/datasources")
	err := g.QueryAndDecode(url, &datasources)
	return datasources, err
}

func (g *Grafana) GetGrafanaAlertingRules() ([]types.GrafanaAlertGroup, error) {
	rules := types.GrafanaAlertRulesResponse{}
	url := g.RelativeLink("/api/prometheus/grafana/api/v1/rules")
	err := g.QueryAndDecode(url, &rules)
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}

func (g *Grafana) GetDatasourceAlertingRules(datasourceUID string) ([]types.GrafanaAlertGroup, error) {
	rules := types.GrafanaAlertRulesResponse{}
	url := g.RelativeLink(fmt.Sprintf("/api/prometheus/%s/api/v1/rules", datasourceUID))
	err := g.QueryAndDecode(url, &rules)
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}

func (g *Grafana) GetPrometheusAlertingRules() ([]types.GrafanaAlertGroup, error) {
	datasources, err := g.GetDatasources()
	if err != nil {
		return nil, err
	}

	groups := []types.GrafanaAlertGroup{}
	for _, ds := range datasources {
		if ds.Type == "prometheus" {
			resp, err := g.GetDatasourceAlertingRules(ds.UID)
			if err != nil {
				return nil, err
			}

			groups = append(groups, resp...)
		}
	}

	return groups, err
}

func (g *Grafana) GetAllAlertingRules() ([]types.GrafanaAlertGroup, error) {
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

func (g *Grafana) CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error) {
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	res := types.SilenceCreateResponse{}
	err := g.QueryAndDecodePost(url, silence, &res)
	return res, err
}

func (g *Grafana) GetSilences() ([]types.Silence, error) {
	silences := []types.Silence{}
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	err := g.QueryAndDecode(url, &silences)
	return silences, err
}

func (g *Grafana) GetSilence(silenceID string) (types.Silence, error) {
	silence := types.Silence{}
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silence/" + silenceID)
	err := g.QueryAndDecode(url, &silence)
	return silence, err
}

func (g *Grafana) DeleteSilence(silenceID string) error {
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silence/" + silenceID)
	return g.QueryDelete(url)
}

/* Query functions */

func (g *Grafana) Query(url string) (io.ReadCloser, error) {
	return g.DoQuery("GET", url, nil)
}

func (g *Grafana) QueryPost(url string, body interface{}) (io.ReadCloser, error) {
	return g.DoQuery("POST", url, body)
}

func (g *Grafana) QueryDelete(url string) error {
	_, err := g.DoQuery("DELETE", url, nil)
	return err
}

func (g *Grafana) QueryAndDecode(url string, output interface{}) error {
	body, err := g.Query(url)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (g *Grafana) QueryAndDecodePost(url string, postBody interface{}, output interface{}) error {
	body, err := g.QueryPost(url, postBody)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (g *Grafana) DoQuery(method string, url string, body interface{}) (io.ReadCloser, error) {
	client := &http.Client{}

	var req *http.Request
	var err error

	if body != nil {
		buffer := new(bytes.Buffer)

		if err := json.NewEncoder(buffer).Encode(body); err != nil {
			return nil, err
		}

		req, err = http.NewRequest(method, url, buffer)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	g.Logger.Trace().
		Str("url", url).
		Str("method", method).
		Msg("Doing a Grafana API query")

	if g.UseAuth() {
		req.SetBasicAuth(g.Config.User, g.Config.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		g.Logger.Error().
			Str("url", url).
			Str("method", method).
			Err(err).
			Msg("Error querying Grafana")
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		g.Logger.Error().
			Str("url", url).
			Str("method", method).
			Int("status", resp.StatusCode).
			Msg("Got error code from Grafana")
		return nil, fmt.Errorf("Could not fetch request. Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
