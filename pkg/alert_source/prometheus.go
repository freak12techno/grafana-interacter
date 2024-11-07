package alert_source

import (
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"

	"github.com/rs/zerolog"
)

type Prometheus struct {
	Config *config.PrometheusConfig
	Logger zerolog.Logger
	Client *http.Client
}

func InitPrometheus(config *config.PrometheusConfig, logger *zerolog.Logger) *Prometheus {
	return &Prometheus{
		Config: config,
		Logger: logger.With().Str("component", "prometheus").Logger(),
		Client: http.NewClient(logger, "prometheus"),
	}
}

func (p *Prometheus) Enabled() bool {
	return p.Config != nil
}

func (p *Prometheus) Name() string {
	return "Prometheus"
}

func (p *Prometheus) GetAuth() *http.Auth {
	if p.Config == nil || p.Config.User == "" || p.Config.Password == "" {
		return nil
	}

	return &http.Auth{Username: p.Config.User, Password: p.Config.Password}
}

func (p *Prometheus) GetAlertingRules() (types.GrafanaAlertGroups, error) {
	if !p.Enabled() {
		return types.GrafanaAlertGroups{}, nil
	}

	rules := types.GrafanaAlertRulesResponse{}
	err := p.Client.Get(p.Config.URL+"/api/v1/rules", &rules, p.GetAuth())
	if err != nil {
		return nil, err
	}

	return rules.Data.Groups, nil
}
