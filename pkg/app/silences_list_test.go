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
func TestAppSilencesListNoSilenceManagers(t *testing.T) {
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
		types.TelegramResponseHasText("No silence managers configured!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseSilenceManagerForListSilences(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppSilencesMultipleSilenceManagers(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
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
			"Choose a silence manager to get silences from:",
			types.TelegramInlineKeyboardResponse{
				InlineKeyboard: [][]types.TelegramInlineKeyboard{
					{
						{
							Unique:       "grafana_paginated_silences_list_",
							Text:         "Grafana",
							CallbackData: "\fgrafana_paginated_silences_list_|0",
						},
					},
					{
						{
							Unique:       "alertmanager_paginated_silences_list_",
							Text:         "Alertmanager",
							CallbackData: "\falertmanager_paginated_silences_list_|0",
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
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseSilenceManagerForListSilences(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppSilencesSilenceManagerFail(t *testing.T) {
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
		"http://alertmanager.com/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error fetching silences: Get \"http://alertmanager.com/api/v2/silences\": custom error\n"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseSilenceManagerForListSilences(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppSilencesLastPage(t *testing.T) {
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
		"http://alertmanager.com/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"=~http://alertmanager.com/api/v2/alerts.*",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-alerts.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/silences-ok.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.PrometheusPaginatedFiringAlertsList,
			Data:   "1",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/silences",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListSilencesFromCallback(app.AlertSourcesWithSilenceManager[1].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppSilencesListFirstPage(t *testing.T) {
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
		"http://alertmanager.com/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"=~http://alertmanager.com/api/v2/alerts.*",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-alerts.json")))

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
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleChooseSilenceManagerForListSilences(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppListSilencesAlertInvalidCallback(t *testing.T) {
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
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.AlertmanagerPaginatedSilencesList,
			Data:   "not-a-number",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/silences",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListSilencesFromCallback(app.AlertSourcesWithSilenceManager[1].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppListSilenceManagerSilenceManagerDisabled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(false)},
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
		types.TelegramResponseHasText("Alertmanager is disabled."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.AlertmanagerPaginatedSilencesList,
			Data:   "0",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/silences",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListSilences(app.AlertSourcesWithSilenceManager[1].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppListSilencesNoSilences(t *testing.T) {
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
		"http://alertmanager.com/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("empty-array.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("<strong>Silences</strong>\nNo silences."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/silences",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.AlertmanagerPaginatedSilencesList,
			Data:   "0",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/silences",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleListSilences(app.AlertSourcesWithSilenceManager[1].SilenceManager)(ctx)
	require.NoError(t, err)
}
