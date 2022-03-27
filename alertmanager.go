package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

type AlertmanagerStruct struct {
	Config *AlertmanagerConfig
	Logger zerolog.Logger
}

func InitAlertmanager(config *AlertmanagerConfig, logger *zerolog.Logger) *AlertmanagerStruct {
	return &AlertmanagerStruct{
		Config: config,
		Logger: logger.With().Str("component", "alertmanager").Logger(),
	}
}

func (g *AlertmanagerStruct) CreateSilence(silence Silence) error {
	url := g.RelativeLink("/api/v2/silences")
	res := Silence{}
	err := g.QueryAndDecodePost(url, silence, res)
	return err
}

func (g *AlertmanagerStruct) GetSilences() ([]Silence, error) {
	silences := []Silence{}
	url := g.RelativeLink("/api/v2/silences")
	err := g.QueryAndDecode(url, &silences)
	return silences, err
}

func (g *AlertmanagerStruct) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
}

func (g *AlertmanagerStruct) QueryPost(url string, body interface{}) (io.ReadCloser, error) {
	if g.Config == nil || g.Config.Password == "" || g.Config.User == "" {
		return nil, fmt.Errorf("Alertmanager API not configured")
	}

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

	g.Logger.Trace().Str("url", url).Msg("Doing an Alertmanager API query")

	req.SetBasicAuth(g.Config.User, g.Config.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Could not fetch request. Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (g *AlertmanagerStruct) QueryAndDecodePost(url string, postBody interface{}, output interface{}) error {
	body, err := g.QueryPost(url, postBody)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (g *AlertmanagerStruct) Query(url string) (io.ReadCloser, error) {
	if g.Config == nil || g.Config.Password == "" && g.Config.User == "" {
		return nil, fmt.Errorf("Alertmanager API not configured")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	g.Logger.Trace().Str("url", url).Msg("Doing a Grafana API query")

	req.SetBasicAuth(g.Config.User, g.Config.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Could not fetch request. Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (g *AlertmanagerStruct) QueryAndDecode(url string, output interface{}) error {
	body, err := g.Query(url)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}
