package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
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

func (g *GrafanaStruct) RenderPanel(panel *PanelStruct) (io.ReadCloser, error) {
	from := time.Now().Unix() * 1000
	to := time.Now().Add(-30*time.Minute).Unix() * 1000

	url := fmt.Sprintf(
		"%s/render/d-solo/%s/dashboard?orgId=1&from=%d&to=%d&panelId=%s&width=1000&height=500&tz=Europe/Moscow",
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
