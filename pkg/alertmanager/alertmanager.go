package alertmanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"main/pkg/config"
	"main/pkg/types"
	"net/http"

	"github.com/rs/zerolog"
)

type Alertmanager struct {
	Config config.AlertmanagerConfig
	Logger zerolog.Logger
}

func InitAlertmanager(config config.AlertmanagerConfig, logger *zerolog.Logger) *Alertmanager {
	return &Alertmanager{
		Config: config,
		Logger: logger.With().Str("component", "alertmanager").Logger(),
	}
}

func (g *Alertmanager) Enabled() bool {
	return g.Config.User != "" && g.Config.Password != ""
}

func (g *Alertmanager) CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error) {
	url := g.RelativeLink("/api/v2/silences")
	res := types.SilenceCreateResponse{}
	err := g.QueryAndDecodePost(url, silence, &res)
	return res, err
}

func (g *Alertmanager) GetSilenceMatchingAlerts(silence types.Silence) ([]types.AlertmanagerAlert, error) {
	relativeUrl := fmt.Sprintf(
		"/api/v2/alerts?%s&silenced=true&inhibited=true&active=true",
		silence.GetFilterQueryString(),
	)
	url := g.RelativeLink(relativeUrl)
	var res []types.AlertmanagerAlert
	err := g.QueryAndDecode(url, &res)
	return res, err
}

func (g *Alertmanager) GetSilences() (types.Silences, error) {
	silences := types.Silences{}
	url := g.RelativeLink("/api/v2/silences")
	err := g.QueryAndDecode(url, &silences)
	return silences, err
}

func (g *Alertmanager) GetSilence(silenceID string) (types.Silence, error) {
	silence := types.Silence{}
	url := g.RelativeLink("/api/v2/silence/" + silenceID)
	err := g.QueryAndDecode(url, &silence)
	return silence, err
}

func (g *Alertmanager) DeleteSilence(silenceID string) error {
	url := g.RelativeLink("/api/v2/silence/" + silenceID)
	return g.QueryDelete(url)
}

/* Helpers */

func (g *Alertmanager) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
}

func (g *Alertmanager) GetSilenceURL(silence types.Silence) string {
	return fmt.Sprintf("%s/#/silences/%s", g.Config.URL, silence.ID)
}

/* Query functions */

func (g *Alertmanager) Query(url string) (io.ReadCloser, error) {
	return g.DoQuery("GET", url, nil)
}

func (g *Alertmanager) QueryDelete(url string) error {
	_, err := g.DoQuery("DELETE", url, nil)
	return err
}

func (g *Alertmanager) QueryPost(url string, body interface{}) (io.ReadCloser, error) {
	return g.DoQuery("POST", url, body)
}

func (g *Alertmanager) QueryAndDecode(url string, output interface{}) error {
	body, err := g.Query(url)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (g *Alertmanager) QueryAndDecodePost(url string, postBody interface{}, output interface{}) error {
	body, err := g.QueryPost(url, postBody)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (g *Alertmanager) DoQuery(method string, url string, body interface{}) (io.ReadCloser, error) {
	if !g.Enabled() {
		return nil, errors.New("Alertmanager API not configured")
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
		g.Logger.Error().
			Str("url", url).
			Str("method", method).
			Err(err).
			Msg("Error querying Alertmanager")
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		g.Logger.Error().
			Str("url", url).
			Str("method", method).
			Int("status", resp.StatusCode).
			Msg("Got error code from Alertmanager")
		return nil, fmt.Errorf("Could not fetch request. Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
