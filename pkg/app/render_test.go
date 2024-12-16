package app

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fs"
	"main/pkg/types"
	"testing"

	"github.com/guregu/null/v5"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestAppRenderPanelErrorFetchingPanels(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error querying for panels: Get \"https://example.com/api/search?type=dash-db\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render PanelName",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleRenderPanel(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelPanelNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok-single.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/alertmanager",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Could not find a panel. See /dashboards for dashboards list, and /dashboard <dashboard name> for its panels."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render asdasdanbxcascasd",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleRenderPanel(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelRenderError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok-single.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/alertmanager",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/render/d-solo/alertmanager/dashboard?panelId=26",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error rendering panel: Get \"https://example.com/render/d-solo/alertmanager/dashboard?panelId=26\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render versions",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleRenderPanel(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok-single.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/alertmanager",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/render/d-solo/alertmanager/dashboard?panelId=26",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("render.jpeg")))

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendPhoto",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render versions",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleRenderPanel(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelChooseDashboardErrorFetchingDashboards(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error fetching dashboards: Get \"https://example.com/api/search?type=dash-db\": custom error\n"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewErrorResponder(errors.New("custom error")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleRenderPanel(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChooseDashboardInvalidInvocation(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Failed to parse page number from callback!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChooseDashboardPrefix,
			Data:   "asd",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderChooseDashboardFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChooseDashboardNoDashboards(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("empty-array.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("No dashboards configured!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChooseDashboardPrefix,
			Data:   "1",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderChooseDashboardFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChooseDashboardOkFromCallback(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		types.TelegramResponseHasText("Choose a dashboard to render (6 - 10 of 28):"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChooseDashboardPrefix,
			Data:   "1",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderChooseDashboardFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChooseDashboardOkFromCommand(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Choose a dashboard to render (1 - 5 of 28):"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleRenderPanel(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChoosePanelWrongCallbackData(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Invalid callback provided!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChoosePanelPrefix,
			Data:   "aaa",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelChoosePanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChoosePanelInvalidPage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Failed to parse page number from callback!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChoosePanelPrefix,
			Data:   "aaa aaa",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelChoosePanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChoosePanelFailedToFetchDashboard(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error fetching dashboard: Get \"https://example.com/api/dashboards/uid/dashboard\": custom error\n"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewErrorResponder(errors.New("custom error")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChoosePanelPrefix,
			Data:   "dashboard 0",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelChoosePanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChoosePanelDashboardWithoutPanels(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("This dashboard has no panels!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("empty.json")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChoosePanelPrefix,
			Data:   "dashboard 0",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelChoosePanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderChoosePanelOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		types.TelegramResponseHasText("Dashboard: Alertmanager\nChoose a panel to render (6 - 10 of 11):"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderChoosePanelPrefix,
			Data:   "dashboard 1",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelChoosePanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelFromCallbackInvalidInvocation(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Invalid callback provided!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderRenderPanelPrefix,
			Data:   "asd",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelFromCallbackErrorFetchingDashboard(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error fetching dashboard: Get \"https://example.com/api/dashboards/uid/dashboard\": custom error\n"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewErrorResponder(errors.New("custom error")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderRenderPanelPrefix,
			Data:   "dashboard 123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelFromCallbackPanelNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Panel not found!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderRenderPanelPrefix,
			Data:   "dashboard 123123123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelFromCallbackFailedToRender(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error rendering panel: Get \"https://example.com/render/d-solo/alertmanager/dashboard?panelId=123\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok-single.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/alertmanager",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/render/d-solo/alertmanager/dashboard?panelId=123",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendPhoto",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
			ReplyTo: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderRenderPanelPrefix,
			Data:   "dashboard 123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelFromCallback(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppRenderPanelFromCallbackOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/dashboard",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/search?type=dash-db",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok-single.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/dashboards/uid/alertmanager",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboard-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/render/d-solo/alertmanager/dashboard?panelId=123",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("render.jpeg")))

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendPhoto",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/render",
			Chat:   &tele.Chat{ID: 2},
			ReplyTo: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaRenderRenderPanelPrefix,
			Data:   "dashboard 123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/render",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleRenderPanelFromCallback(ctx)
	require.NoError(t, err)
}
