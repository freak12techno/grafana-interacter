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

func (g *AlertmanagerStruct) CreateSilence(silence Silence) (Silence, error) {
	url := g.RelativeLink("/api/v2/silences")
	res := Silence{}
	err := g.QueryAndDecodePost(url, silence, res)
	return silence, err
}

func (g *AlertmanagerStruct) GetSilences() ([]Silence, error) {
	silences := []Silence{}
	url := g.RelativeLink("/api/v2/silences")
	err := g.QueryAndDecode(url, &silences)
	return silences, err
}

func (g *AlertmanagerStruct) DeleteSilence(silenceID string) error {
	url := g.RelativeLink("/api/v2/silence/" + silenceID)
	return g.QueryDelete(url)
}

/* Helpers */

func (g *AlertmanagerStruct) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
}

func (g *AlertmanagerStruct) GetSilenceURL(silence Silence) string {
	return fmt.Sprintf("%s/#/silences/%s", g.Config.URL, silence.ID)
}

/* Query functions */

func (g *AlertmanagerStruct) Query(url string) (io.ReadCloser, error) {
	return g.DoQuery("GET", url, nil)
}

func (g *AlertmanagerStruct) QueryDelete(url string) error {
	_, err := g.DoQuery("DELETE", url, nil)
	return err
}

func (g *AlertmanagerStruct) QueryPost(url string, body interface{}) (io.ReadCloser, error) {
	return g.DoQuery("POST", url, body)
}

func (g *AlertmanagerStruct) QueryAndDecode(url string, output interface{}) error {
	body, err := g.Query(url)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (g *AlertmanagerStruct) QueryAndDecodePost(url string, postBody interface{}, output interface{}) error {
	body, err := g.QueryPost(url, postBody)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (g *AlertmanagerStruct) DoQuery(method string, url string, body interface{}) (io.ReadCloser, error) {
	if g.Config == nil || g.Config.Password == "" || g.Config.User == "" {
		return nil, fmt.Errorf("Alertmanager API not configured")
	}

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
		Msg("Doing an Alertmanager API query")

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
