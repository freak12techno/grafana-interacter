package app

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/types"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestAppRenderPanelInvalidInvocation(t *testing.T) {
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
		types.TelegramResponseHasText("Usage: /render <opts> <panel name>"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
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

	app := NewApp(config, "1.2.3")
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
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

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

	app := NewApp(config, "1.2.3")
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
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

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

	app := NewApp(config, "1.2.3")
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
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("grafana-dashboards-ok.json")))

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

	app := NewApp(config, "1.2.3")
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
