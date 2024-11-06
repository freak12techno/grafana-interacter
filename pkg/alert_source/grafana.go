package alert_source

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"

	"github.com/rs/zerolog"
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

func (g *Grafana) Name() string {
	return "Grafana"
}

func (g *Grafana) Enabled() bool {
	return g.Config.Alerts.Bool
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

func (g *Grafana) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
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
