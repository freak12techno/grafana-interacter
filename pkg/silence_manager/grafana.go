package silence_manager

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
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

func (g *Grafana) Prefixes() Prefixes {
	return Prefixes{
		PaginatedSilencesList: constants.GrafanaPaginatedSilencesList,
		Silence:               constants.GrafanaSilencePrefix,
		PrepareSilence:        constants.GrafanaPrepareSilencePrefix,
		Unsilence:             constants.GrafanaUnsilencePrefix,
		ListSilencesCommand:   constants.GrafanaListSilencesCommand,
	}
}

func (g *Grafana) Name() string {
	return "Grafana"
}

func (g *Grafana) Enabled() bool {
	return g.Config.Silences.Bool
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

func (g *Grafana) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
}

func (g *Grafana) CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error) {
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	res := types.SilenceCreateResponse{}
	err := g.Client.Post(url, silence, &res, g.GetAuth())
	return res, err
}

func (g *Grafana) GetSilences() (types.Silences, error) {
	silences := types.Silences{}
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silences")
	err := g.Client.Get(url, &silences, g.GetAuth())
	return silences, err
}

func (g *Grafana) GetSilence(silenceID string) (types.Silence, error) {
	silence := types.Silence{}
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silence/" + silenceID)
	err := g.Client.Get(url, &silence, g.GetAuth())
	return silence, err
}

func (g *Grafana) DeleteSilence(silenceID string) error {
	url := g.RelativeLink("/api/alertmanager/grafana/api/v2/silence/" + silenceID)
	return g.Client.Delete(url, g.GetAuth())
}

func (g *Grafana) GetSilenceMatchingAlerts(silence types.Silence) ([]types.AlertmanagerAlert, error) {
	relativeUrl := fmt.Sprintf(
		"/api/alertmanager/grafana/api/v2/alerts?%s&silenced=true&inhibited=true&active=true",
		silence.GetFilterQueryString(),
	)
	url := g.RelativeLink(relativeUrl)
	var res []types.AlertmanagerAlert
	err := g.Client.Get(url, &res, g.GetAuth())
	return res, err
}
