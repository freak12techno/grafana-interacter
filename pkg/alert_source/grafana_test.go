package alert_source

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"testing"

	"github.com/guregu/null/v5"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestGrafanaBasic(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{
		URL:            "http://localhost:9090",
		MutesDurations: []string{"1h"},
		Silences:       null.BoolFrom(true),
	}
	client := InitGrafana(config, logger)

	require.True(t, client.Enabled())
	require.Equal(t, "Grafana", client.Name())
}

//nolint:paralleltest
func TestGrafanaGetAlertingRulesFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/prometheus/grafana/api/v1/rules",
		httpmock.NewErrorResponder(errors.New("custom error")))

	rules, err := client.GetAlertingRules()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, rules)
}

//nolint:paralleltest
func TestGrafanaGetAlertingRulesOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/prometheus/grafana/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")))

	rules, err := client.GetAlertingRules()
	require.NoError(t, err)
	require.NotEmpty(t, rules)
}
