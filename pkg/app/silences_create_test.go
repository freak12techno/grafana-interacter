package app

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"testing"

	"github.com/guregu/null/v5"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestAppCreateSilenceSilenceManagerDisabled(t *testing.T) {
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
		types.TelegramResponseHasText("Grafana is disabled."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleNewSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceInvalidInvocation(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
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
		types.TelegramResponseHasText("Usage: /grafana_silence <duration> <params>"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleNewSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceErrorCreatingSilence(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error creating silence: Post \"https://example.com/api/alertmanager/grafana/api/v2/silences\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 48h host=test",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleNewSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceErrorGettingSilence(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-create-silence-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/005a07f4-3e6b-4fc1-b97e-6cb928135281",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error getting created silence: Get \"https://example.com/api/alertmanager/grafana/api/v2/silence/005a07f4-3e6b-4fc1-b97e-6cb928135281\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 48h host=test",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleNewSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceErrorGettingSilenceAlerts(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-create-silence-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/005a07f4-3e6b-4fc1-b97e-6cb928135281",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silence-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=network%3D%22neutron%22&filter=alertname%3D%22CosmosNodeNotLatestBinary%22&silenced=true&inhibited=true&active=true",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error getting alerts for silence: Get \"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=network%3D%22neutron%22&filter=alertname%3D%22CosmosNodeNotLatestBinary%22&silenced=true&inhibited=true&active=true\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 48h host=test",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleNewSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-create-silence-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/005a07f4-3e6b-4fc1-b97e-6cb928135281",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silence-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=network%3D%22neutron%22&filter=alertname%3D%22CosmosNodeNotLatestBinary%22&silenced=true&inhibited=true&active=true",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-alerts.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytesAndMarkup(assets.GetBytesOrPanic("responses/silence-create-ok.html"), types.TelegramInlineKeyboardResponse{
			InlineKeyboard: [][]types.TelegramInlineKeyboard{
				{{
					Unique:       "clear_keyboard_",
					Text:         "✅Confirm",
					CallbackData: "\fclear_keyboard_",
				}},
				{{
					Unique:       "grafana_unsilence_",
					Text:         "❌Unsilence",
					CallbackData: "\fgrafana_unsilence_|4de5faa2-8c0c-4c66-bd31-25c3bf5fa231 1",
				}},
			},
		}),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 48h host=test",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleNewSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppPrepareSilenceViaCallbackErrorGettingSilenceAlerts(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/prometheus/grafana/api/v1/rules",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error querying alerts: Get \"https://example.com/api/prometheus/grafana/api/v1/rules\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandlePrepareNewSilenceFromCallback(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppPrepareSilenceViaCallbackAlertNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
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
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Alert was not found!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandlePrepareNewSilenceFromCallback(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppPrepareSilenceViaCallbackOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:            "https://example.com",
			Silences:       null.BoolFrom(true),
			MutesDurations: []string{"1h", "3h"},
		},
		Alertmanager: nil,
		Prometheus:   nil,
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
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytesAndMarkup(assets.GetBytesOrPanic("responses/silence-prepare-ok.html"), types.TelegramInlineKeyboardResponse{
			InlineKeyboard: [][]types.TelegramInlineKeyboard{
				{
					{
						Unique:       "grafana_silence_",
						Text:         "⌛ Silence for 1h",
						CallbackData: "\fgrafana_silence_|1h 0dd372ce555502a912ed8331b00ff883",
					},
				},
				{
					{
						Unique:       "grafana_silence_",
						Text:         "⌛ Silence for 3h",
						CallbackData: "\fgrafana_silence_|3h 0dd372ce555502a912ed8331b00ff883",
					},
				},
			},
		}),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "0dd372ce555502a912ed8331b00ff883",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandlePrepareNewSilenceFromCallback(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceViaCallbackInvalidPayload(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
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
		types.TelegramResponseHasText("Invalid callback provided!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 4",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence 4",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleCallbackNewSilence(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceViaCallbackErrorFetchingAlerts(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/prometheus/grafana/api/v1/rules",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error querying alerts: Get \"https://example.com/api/prometheus/grafana/api/v1/rules\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 4",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "invalid 123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleCallbackNewSilence(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceViaCallbackInvalidDuration(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
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
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Invalid duration provided!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 4",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "invalid 123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleCallbackNewSilence(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceViaCallbackAlertNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
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
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Alert was not found!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 4",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "48h 123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleCallbackNewSilence(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppCreateSilenceViaCallbackOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", Silences: null.BoolFrom(true)},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/prometheus/grafana/api/v1/rules",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("prometheus-alerting-rules-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error creating silence: Post \"https://example.com/api/alertmanager/grafana/api/v2/silences\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_silence 4",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaSilencePrefix,
			Data:   "48h 0dd372ce555502a912ed8331b00ff883",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_silence",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleCallbackNewSilence(
		app.AlertSourcesWithSilenceManager[0].SilenceManager,
		app.AlertSourcesWithSilenceManager[0].AlertSource,
	)(ctx)
	require.NoError(t, err)
}
