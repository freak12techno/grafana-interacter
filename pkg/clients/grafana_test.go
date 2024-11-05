package clients

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/types"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestGrafanaBasic(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "http://localhost:9090", MutesDurations: []string{"1h"}}
	client := InitGrafana(config, logger)

	require.True(t, client.Enabled())
	require.Equal(t, "Grafana", client.Name())
	require.Equal(t, constants.GrafanaSilencePrefix, client.GetSilencePrefix())
	require.Equal(t, constants.GrafanaUnsilencePrefix, client.GetUnsilencePrefix())
	require.Equal(t, constants.GrafanaPaginatedSilencesList, client.GetPaginatedSilencesListPrefix())
	require.Equal(t, []string{"1h"}, client.GetMutesDurations())
}

//nolint:paralleltest
func TestGrafanaGetAllDashboardsFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewErrorResponder(errors.New("custom error")))

	dashboards, err := client.GetAllDashboards()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, dashboards)
}

//nolint:paralleltest
func TestGrafanaGetAllDashboardsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

	dashboards, err := client.GetAllDashboards()
	require.NoError(t, err)
	require.NotEmpty(t, dashboards)
}

//nolint:paralleltest
func TestGrafanaGetDashboardFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewErrorResponder(errors.New("custom error")))

	dashboard, err := client.GetDashboard("dashboard")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, dashboard)
}

//nolint:paralleltest
func TestGrafanaGetDashboardOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	dashboard, err := client.GetDashboard("dashboard")
	require.NoError(t, err)
	require.NotEmpty(t, dashboard)
}

//nolint:paralleltest
func TestGrafanaGetAllPanelsDashboardFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewErrorResponder(errors.New("custom error")))

	panels, err := client.GetAllPanels()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, panels)
}

//nolint:paralleltest
func TestGrafanaGetAllPanelsPanelFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/alertmanager",
		httpmock.NewErrorResponder(errors.New("custom error")))

	panels, err := client.GetAllPanels()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, panels)
}

//nolint:paralleltest
func TestGrafanaGetAllPanelsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/alertmanager",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	panels, err := client.GetAllPanels()
	require.NoError(t, err)
	require.NotEmpty(t, panels)
}

//nolint:paralleltest
func TestGrafanaGetDatasourcesFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/datasources",
		httpmock.NewErrorResponder(errors.New("custom error")))

	datasources, err := client.GetDatasources()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, datasources)
}

//nolint:paralleltest
func TestGrafanaGetDatasourcesOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/datasources",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-datasources-ok.json")))

	datasources, err := client.GetDatasources()
	require.NoError(t, err)
	require.NotEmpty(t, datasources)
}

//nolint:paralleltest
func TestGrafanaGetAlertingRulesFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
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

//nolint:paralleltest
func TestGrafanaCreateSilenceFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", Token: "token"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	_, err := client.CreateSilence(types.Silence{})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest
func TestGrafanaCreateSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-create-silence-ok.json")))

	silence, err := client.CreateSilence(types.Silence{ID: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, silence)
}

//nolint:paralleltest
func TestGrafanaGetSilencesFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	silences, err := client.GetSilences()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, silences)
}

//nolint:paralleltest
func TestGrafanaGetSilencesOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	silences, err := client.GetSilences()
	require.NoError(t, err)
	require.NotEmpty(t, silences)
}

//nolint:paralleltest
func TestGrafanaGetSilenceFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/id",
		httpmock.NewErrorResponder(errors.New("custom error")))

	silence, err := client.GetSilence("id")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, silence)
}

//nolint:paralleltest
func TestGrafanaGetSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/id",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silence-ok.json")))

	silence, err := client.GetSilence("id")
	require.NoError(t, err)
	require.NotEmpty(t, silence)
}

//nolint:paralleltest
func TestGrafanaDeleteSilenceFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/id",
		httpmock.NewErrorResponder(errors.New("custom error")))

	err := client.DeleteSilence("id")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest
func TestGrafanaDeleteSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/id",
		httpmock.NewBytesResponder(200, []byte{}))

	err := client.DeleteSilence("id")
	require.NoError(t, err)
}

//nolint:paralleltest
func TestGrafanaListSilenceAlertsFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=key%3D%22value%22&silenced=true&inhibited=true&active=true",
		httpmock.NewErrorResponder(errors.New("custom error")))

	alerts, err := client.GetSilenceMatchingAlerts(types.Silence{
		Matchers: types.SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, alerts)
}

//nolint:paralleltest
func TestGrafanaListSilenceAlertsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=key%3D%22value%22&silenced=true&inhibited=true&active=true",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-alerts.json")))

	alerts, err := client.GetSilenceMatchingAlerts(types.Silence{
		Matchers: types.SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, alerts)
}

//nolint:paralleltest
func TestGrafanaRenderPanelFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/render/d-solo/dashboard/dashboard?panelId=1",
		httpmock.NewErrorResponder(errors.New("custom error")))

	alerts, err := client.RenderPanel(&types.PanelStruct{
		DashboardID: "dashboard",
		PanelID:     1,
	}, map[string]string{})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, alerts)
}

//nolint:paralleltest
func TestGrafanaRenderPanelOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com"}
	client := InitGrafana(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/render/d-solo/dashboard/dashboard?panelId=1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("render.jpeg")))

	render, err := client.RenderPanel(&types.PanelStruct{
		DashboardID: "dashboard",
		PanelID:     1,
	}, map[string]string{})
	defer func() {
		_ = render.Close()
	}()

	require.NoError(t, err)
	require.NotNil(t, render)
}
