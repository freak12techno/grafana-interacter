package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/pkg/config"
	"main/pkg/types"
	"net/http"

	"github.com/rs/zerolog"
)

type Prometheus struct {
	Config *config.PrometheusConfig
	Logger zerolog.Logger
}

func InitPrometheus(config *config.PrometheusConfig, logger *zerolog.Logger) *Prometheus {
	return &Prometheus{
		Config: config,
		Logger: logger.With().Str("component", "prometheus").Logger(),
	}
}

func (p *Prometheus) UseAuth() bool {
	return p.Config.User != "" && p.Config.Password != ""
}

func (p *Prometheus) Enabled() bool {
	return p.Config != nil
}

func (p *Prometheus) Name() string {
	return "Prometheus"
}

func (p *Prometheus) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", p.Config.URL, url)
}

func (p *Prometheus) GetAlertingRules() (types.GrafanaAlertGroups, error) {
	if !p.Enabled() {
		return types.GrafanaAlertGroups{}, nil
	}

	rules := types.GrafanaAlertRulesResponse{}
	url := p.RelativeLink("/api/v1/rules")
	err := p.QueryAndDecode(url, &rules)
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}

/* Query functions */

func (p *Prometheus) Query(url string) (io.ReadCloser, error) {
	return p.DoQuery("GET", url, nil)
}

func (p *Prometheus) QueryAndDecode(url string, output interface{}) error {
	body, err := p.Query(url)
	if err != nil {
		return err
	}

	defer body.Close()
	return json.NewDecoder(body).Decode(&output)
}

func (p *Prometheus) DoQuery(method string, url string, body interface{}) (io.ReadCloser, error) {
	var transport http.RoundTripper

	transportRaw, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		transport = transportRaw.Clone()
	} else {
		transport = http.DefaultTransport
	}

	client := &http.Client{
		Transport: transport,
	}

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

	p.Logger.Trace().
		Str("url", url).
		Str("method", method).
		Msg("Doing a Prometheus API query")

	if p.UseAuth() {
		req.SetBasicAuth(p.Config.User, p.Config.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		p.Logger.Error().
			Str("url", url).
			Str("method", method).
			Err(err).
			Msg("Error querying Prometheus")
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		p.Logger.Error().
			Str("url", url).
			Str("method", method).
			Int("status", resp.StatusCode).
			Msg("Got error code from Prometheus")
		return nil, fmt.Errorf("Could not fetch request. Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
