package clients

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"main/pkg/config"
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

func (g *Grafana) RenderPanel(
	panelID int,
	dashboardID string,
	qs map[string]string,
) (io.ReadCloser, error) {
	params := generic.MergeMaps(g.Config.RenderOptions, qs)
	params["panelId"] = strconv.Itoa(panelID)

	url := g.RelativeLink(fmt.Sprintf(
		"/render/d-solo/%s/dashboard?%s",
		dashboardID,
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
