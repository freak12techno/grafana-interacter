package clients

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestGrafanaGetAllDashboardsFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitGrafana(config, logger)
	require.Empty(t, client.GetMutesDurations())

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
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok-single.json")))

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
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok-single.json")))

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

	render, err := client.RenderPanel(1, "dashboard", map[string]string{})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, render)
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

	render, err := client.RenderPanel(1, "dashboard", map[string]string{})
	defer func() {
		_ = render.Close()
	}()

	require.NoError(t, err)
	require.NotNil(t, render)
}
