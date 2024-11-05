package clients

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/http"
	"main/pkg/types"

	"github.com/rs/zerolog"
)

type Alertmanager struct {
	Config *config.AlertmanagerConfig
	Logger zerolog.Logger
	Client *http.Client
}

func InitAlertmanager(config *config.AlertmanagerConfig, logger *zerolog.Logger) *Alertmanager {
	return &Alertmanager{
		Config: config,
		Logger: logger.With().Str("component", "alertmanager").Logger(),
		Client: http.NewClient(logger, "alertmanager"),
	}
}

func (g *Alertmanager) Enabled() bool {
	return g.Config != nil
}

func (g *Alertmanager) Name() string {
	return "Alertmanager"
}

func (g *Alertmanager) GetUnsilencePrefix() string {
	return constants.AlertmanagerUnsilencePrefix
}

func (g *Alertmanager) GetSilencePrefix() string {
	return constants.AlertmanagerSilencePrefix
}

func (g *Alertmanager) GetPaginatedSilencesListPrefix() string {
	return constants.AlertmanagerPaginatedSilencesList
}

func (g *Alertmanager) GetMutesDurations() []string {
	return g.Config.MutesDurations
}

func (g *Alertmanager) GetAuth() *http.Auth {
	if g.Config == nil || g.Config.User == "" || g.Config.Password == "" {
		return nil
	}

	return &http.Auth{Username: g.Config.User, Password: g.Config.Password}
}

func (g *Alertmanager) CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error) {
	url := g.RelativeLink("/api/v2/silences")
	res := types.SilenceCreateResponse{}
	err := g.Client.Post(url, silence, &res, g.GetAuth())
	return res, err
}

func (g *Alertmanager) GetSilenceMatchingAlerts(silence types.Silence) ([]types.AlertmanagerAlert, error) {
	relativeUrl := fmt.Sprintf(
		"/api/v2/alerts?%s&silenced=true&inhibited=true&active=true",
		silence.GetFilterQueryString(),
	)
	url := g.RelativeLink(relativeUrl)
	var res []types.AlertmanagerAlert
	err := g.Client.Get(url, &res, g.GetAuth())
	return res, err
}

func (g *Alertmanager) GetSilences() (types.Silences, error) {
	silences := types.Silences{}
	url := g.RelativeLink("/api/v2/silences")
	err := g.Client.Get(url, &silences, g.GetAuth())
	return silences, err
}

func (g *Alertmanager) GetSilence(silenceID string) (types.Silence, error) {
	silence := types.Silence{}
	url := g.RelativeLink("/api/v2/silence/" + silenceID)
	err := g.Client.Get(url, &silence, g.GetAuth())
	return silence, err
}

func (g *Alertmanager) DeleteSilence(silenceID string) error {
	url := g.RelativeLink("/api/v2/silence/" + silenceID)
	return g.Client.Delete(url, g.GetAuth())
}

func (g *Alertmanager) RelativeLink(url string) string {
	return fmt.Sprintf("%s%s", g.Config.URL, url)
}
