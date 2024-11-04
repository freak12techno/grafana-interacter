package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

type Auth struct {
	Username string
	Password string
}

type Client struct {
	Logger  zerolog.Logger
	Querier string
}

func NewClient(logger *zerolog.Logger, querier string) *Client {
	return &Client{
		Logger: logger.With().
			Str("component", "http").
			Str("querier", querier).
			Logger(),
		Querier: querier,
	}
}

func (c *Client) Get(
	url string,
	target interface{},
	auth *Auth,
) error {
	return c.doQuery(http.MethodGet, url, nil, auth, target)
}

func (c *Client) doQuery(
	method string,
	url string,
	body io.Reader,
	auth *Auth,
	target interface{},
) error {
	var transport http.RoundTripper

	transportRaw, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		transport = transportRaw.Clone()
	} else {
		transport = http.DefaultTransport
	}

	client := &http.Client{Transport: transport}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "grafana-interacter")

	if auth != nil {
		req.SetBasicAuth(auth.Username, auth.Password)
	}

	c.Logger.Debug().
		Str("url", url).
		Str("method", method).
		Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		c.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		c.Logger.Error().
			Str("url", url).
			Str("method", method).
			Int("status", res.StatusCode).
			Msg("Got error code")
		return fmt.Errorf("Could not fetch request. Status code: %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(target)
}
