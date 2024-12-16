package app

import (
	"encoding/json"
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fs"
	"main/pkg/types"
	"main/pkg/types/render"
	"testing"
	"time"

	"github.com/guregu/null/v5"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestAppFiringAlertsNoAlertSources(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:    "https://example.com",
			Alerts: null.BoolFrom(false),
		},
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
		types.TelegramResponseHasText("No alert sources configured!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseAlertSourceForListFiringAlerts(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppFiringAlertsMultipleAlertSources(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(true)},
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
		types.TelegramResponseHasTextAndMarkup(
			"Choose an alert source to get alerts from:",
			types.TelegramInlineKeyboardResponse{
				InlineKeyboard: [][]types.TelegramInlineKeyboard{
					{
						{
							Unique:       "grafana_paginated_firing_alerts_list_",
							Text:         "Grafana",
							CallbackData: "\fgrafana_paginated_firing_alerts_list_|0",
						},
					},
					{
						{
							Unique:       "prometheus_paginated_firing_alerts_list_",
							Text:         "Prometheus",
							CallbackData: "\fprometheus_paginated_firing_alerts_list_|0",
						},
					},
				},
			},
		),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseAlertSourceForListFiringAlerts(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppFiringAlertsAlertSourceFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://prometheus.com/api/v1/rules",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error fetching alerts: Get \"https://prometheus.com/api/v1/rules\": custom error!\n"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseAlertSourceForListFiringAlerts(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppFiringAlertsAlertLastPage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://prometheus.com/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.PrometheusPaginatedFiringAlertsList,
			Data:   "1",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/firing",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListFiringAlertsFromCallback(
		app.AlertSourcesWithSilenceManager[1].AlertSource,
		app.AlertSourcesWithSilenceManager[1].SilenceManager,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppAlertsFiringRenderOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/prometheus/grafana/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/alerts-firing-ok.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.PrometheusPaginatedFiringAlertsList,
			Data:   "1",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/firing",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	alertRulesRaw := assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")
	var alertRules types.GrafanaAlertRulesResponse
	err := json.Unmarshal(alertRulesRaw, &alertRules)
	require.NoError(t, err)

	alerts := alertRules.Data.Groups.FilterFiringOrPendingAlertGroups(true).ToFiringAlerts()
	require.NotEmpty(t, alerts)

	timeParsed, err := time.Parse(time.RFC3339, "2024-11-08T23:34:01Z")
	require.NoError(t, err)

	err = app.EditRender(ctx, "alerts_firing", render.RenderStruct{
		Grafana: app.Grafana,
		Data: types.FiringAlertsListStruct{
			AlertSourceName: "Prometheus",
			Alerts:          alerts[2:],
			AlertsCount:     4,
			Start:           3,
			End:             4,
			RenderTime:      timeParsed,
		},
	})
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppFiringAlertsAlertFirstPage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://prometheus.com/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseAlertSourceForListFiringAlerts(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppFiringAlertsAlertInvalidCallback(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(false)},
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
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.PrometheusPaginatedFiringAlertsList,
			Data:   "not-a-number",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/firing",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListFiringAlertsFromCallback(
		app.AlertSourcesWithSilenceManager[1].AlertSource,
		app.AlertSourcesWithSilenceManager[1].SilenceManager,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppFiringAlertsAlertAlertSourceDisabled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Prometheus is disabled."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.PrometheusPaginatedFiringAlertsList,
			Data:   "0",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/firing",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListFiringAlertsFromCallback(
		app.AlertSourcesWithSilenceManager[1].AlertSource,
		app.AlertSourcesWithSilenceManager[1].SilenceManager,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppFiringAlertsAlertNoAlerts(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Alerts: null.BoolFrom(false)},
		Alertmanager: &configPkg.AlertmanagerConfig{URL: "http://alertmanager.com"},
		Prometheus:   &configPkg.PrometheusConfig{URL: "https://prometheus.com"},
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://prometheus.com/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-empty.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		types.TelegramResponseHasText("No firing alerts."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/firing",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.PrometheusPaginatedFiringAlertsList,
			Data:   "0",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/firing",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListFiringAlertsFromCallback(
		app.AlertSourcesWithSilenceManager[1].AlertSource,
		app.AlertSourcesWithSilenceManager[1].SilenceManager,
	)(ctx)
	require.NoError(t, err)
}
