package clients

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jarcoal/httpmock"
)

func TestPrometheusBasic(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.PrometheusConfig{URL: "http://localhost:9090"}
	client := InitPrometheus(config, logger)

	require.True(t, client.Enabled())
	require.Equal(t, "Prometheus", client.Name())
}

func TestPrometheusGetAlertingRulesDisabled(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := InitPrometheus(nil, logger)

	alertingRules, err := client.GetAlertingRules()
	require.NoError(t, err)
	require.Empty(t, alertingRules)
}

//nolint:paralleltest
func TestPrometheusGetAlertingRulesFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.PrometheusConfig{URL: "https://example.com"}
	client := InitPrometheus(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v1/rules",
		httpmock.NewErrorResponder(errors.New("custom error")))

	alertingRules, err := client.GetAlertingRules()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, alertingRules)
}

//nolint:paralleltest
func TestPrometheusGetAlertingRulesOkWithoutAuth(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.PrometheusConfig{URL: "https://example.com"}
	client := InitPrometheus(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")),
	)

	alertingRules, err := client.GetAlertingRules()
	require.NoError(t, err)
	require.NotEmpty(t, alertingRules)
}

//nolint:paralleltest
func TestPrometheusGetAlertingRulesOkWithAuth(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.PrometheusConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitPrometheus(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")),
	)

	alertingRules, err := client.GetAlertingRules()
	require.NoError(t, err)
	require.NotEmpty(t, alertingRules)
}
