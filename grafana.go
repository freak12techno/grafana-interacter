package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type GrafanaStruct struct {
	URL    string
	Auth   *AuthStruct
	Logger zerolog.Logger
}

type GrafanaDashboardInfo struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type GrafanaDashboardResponse struct {
	Dashboard GrafanaSingleDashboard `json:"dashboard"`
}

type GrafanaSingleDashboard struct {
	Title  string         `json:"title"`
	UID    string         `json:"uid"`
	Panels []GrafanaPanel `json:"panels"`
}

type GrafanaPanel struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
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

func (g *GrafanaStruct) RenderPanel(panel *PanelStruct) (io.ReadCloser, error) {
	from := time.Now().Unix() * 1000
	to := time.Now().Add(-30*time.Minute).Unix() * 1000

	url := fmt.Sprintf(
		"%s/render/d-solo/%s/dashboard?orgId=1&from=%d&to=%d&panelId=%d&width=1000&height=500&tz=Europe/Moscow",
		g.URL,
		panel.DashboardID,
		from,
		to,
		panel.PanelID,
	)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	if g.UseAuth() {
		g.Logger.Trace().Msg("Using basic auth")
		req.SetBasicAuth(g.Auth.User, g.Auth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could not query dashboard: %s", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Could not fetch rendered image. Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (g *GrafanaStruct) GetAllDashboards() ([]GrafanaDashboardInfo, error) {
	url := fmt.Sprintf("%s/api/search?type=dash-db", g.URL)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	if g.UseAuth() {
		g.Logger.Trace().Msg("Using basic auth")
		req.SetBasicAuth(g.Auth.User, g.Auth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dashboards := []GrafanaDashboardInfo{}

	err = json.NewDecoder(resp.Body).Decode(&dashboards)
	if err != nil {
		return nil, err
	}

	return dashboards, nil
}

func (g *GrafanaStruct) GetDashboard(dashboardUID string) (*GrafanaDashboardResponse, error) {
	url := fmt.Sprintf("%s/api/dashboards/uid/%s", g.URL, dashboardUID)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	if g.UseAuth() {
		g.Logger.Trace().Msg("Using basic auth")
		req.SetBasicAuth(g.Auth.User, g.Auth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dashboards := &GrafanaDashboardResponse{}

	err = json.NewDecoder(resp.Body).Decode(&dashboards)
	if err != nil {
		return nil, err
	}

	return dashboards, nil
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

		g.Logger.Info().Int("index", index).Msg("Rendering dashboard")

		group.Go(func() error {
			enrichedDashboard, err := g.GetDashboard(dashboard.UID)
			g.Logger.Info().Int("index", index).Msg("Rendering dashboard")
			if err == nil {
				dashboardsEnriched[index] = *enrichedDashboard
				g.Logger.Info().Int("len", len(enrichedDashboard.Dashboard.Panels)).Msg("Panels length")
			}

			return err
		})
	}

	g.Logger.Info().Int("len", len(dashboards)).Msg("Dashboard length")

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
				PanelID:       p.ID,
			}

			counter++
		}
	}

	return panels, nil
}
