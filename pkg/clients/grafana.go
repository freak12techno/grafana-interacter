package clients

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/http"
	"main/pkg/types"
	"main/pkg/utils"
	"main/pkg/utils/generic"
	"strconv"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Grafana struct {
	Config config.GrafanaConfig
	Logger zerolog.Logger
	Client *http.Client
}

func InitGrafana(config config.GrafanaConfig, logger *zerolog.Logger) *Grafana {
	return &Grafana{
		Config: config,
		Logger: logger.With().Str("component", "grafana").Logger(),
		Client: http.NewClient(logger, "grafana"),
	}
}

func (g *Grafana) GetUnsilencePrefix() string {
	return constants.GrafanaUnsilencePrefix
}

func (g *Grafana) GetSilencePrefix() string {
	return constants.GrafanaSilencePrefix
}

func (g *Grafana) GetPaginatedSilencesListPrefix() string {
	return constants.GrafanaPaginatedSilencesList
}

func (g *Grafana) Name() string {
	return "Grafana"
}

func (g *Grafana) Enabled() bool {
	return true
}

func (g *Grafana) GetMutesDurations() []string {
	return g.Config.MutesDurations
}

func (g *Grafana) GetAuth() *http.Auth {
	if g.Config.User == "" && g.Config.Password == "" && g.Config.Token == "" {
		return nil
	}

	return &http.Auth{
		Username: g.Config.User,
		Password: g.Config.Password,
		Token:    g.Config.Token,
	}
}

func (g *Grafana) RenderPanel(panel *types.PanelStruct, qs map[string]string) (io.ReadCloser, error) {
	params := generic.MergeMaps(g.Config.RenderOptions, qs)
	params["panelId"] = strconv.Itoa(panel.PanelID)

	url := g.RelativeLink(fmt.Sprintf(
		"/render/d-solo/%s/dashboard?%s",
		panel.DashboardID,
		utils.SerializeQueryString(params),
	))

	return g.Client.GetRaw(url, g.GetAuth())
}

func (g *Grafana) GetAllDashboards() (types.GrafanaDashboardsInfo, error) {
	url := g.RelativeLink("/api/search?type=dash-db")
	dashboards := types.GrafanaDashboardsInfo{}
	err := g.Client.Get(url, &dashboards, g.GetAuth())
	return dashboards, err
}

func (g *Grafana) GetDashboard(dashboardUID string) (*types.GrafanaDashboardResponse, error) {
	url := g.RelativeLink("/api/dashboards/uid/" + dashboardUID)
	dashboards := &types.GrafanaDashboardResponse{}
	err := g.Client.Get(url, &dashboards, g.GetAuth())
	return dashboards, err
}

func (g *Grafana) GetAllPanels() (types.PanelsStruct, error) {
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
			enrichedDashboard, dashboardErr := g.GetDashboard(dashboard.UID)
			if dashboardErr == nil {
				dashboardsEnriched[index] = *enrichedDashboard
			}

			return dashboardErr
		})
	}

	if groupErr := group.Wait(); groupErr != nil {
		return nil, groupErr
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
	err := g.Client.Get(url, &datasources, g.GetAuth())
	return datasources, err
}

func (g *Grafana) GetAlertingRules() (types.GrafanaAlertGroups, error) {
	rules := types.GrafanaAlertRulesResponse{}
	url := g.RelativeLink("/api/prometheus/grafana/api/v1/rules")
	err := g.Client.Get(url, &rules, g.GetAuth())
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}

func (g *Grafana) CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error) {
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	res := types.SilenceCreateResponse{}
	err := g.Client.Post(url, silence, &res, g.GetAuth())
	return res, err
}

func (g *Grafana) GetSilences() (types.Silences, error) {
	silences := types.Silences{}
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	err := g.Client.Get(url, &silences, g.GetAuth())
	return silences, err
}

func (g *Grafana) GetSilence(silenceID string) (types.Silence, error) {
	silence := types.Silence{}
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silence/" + silenceID)
	err := g.Client.Get(url, &silence, g.GetAuth())
	return silence, err
}

func (g *Grafana) DeleteSilence(silenceID string) error {
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silence/" + silenceID)
	return g.Client.Delete(url, g.GetAuth())
}

func (g *Grafana) GetSilenceMatchingAlerts(silence types.Silence) ([]types.AlertmanagerAlert, error) {
	relativeUrl := fmt.Sprintf(
		"/api/alertmanager/grafana/api/v2/alerts?%s&silenced=true&inhibited=true&active=true",
		silence.GetFilterQueryString(),
	)
	url := g.RelativeLink(relativeUrl)
	var res []types.AlertmanagerAlert
	err := g.Client.Get(url, &res, g.GetAuth())
	return res, err
}
