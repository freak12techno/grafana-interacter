package app

import (
	"errors"
	"fmt"
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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
func TestAppPrepareSilenceViaCallbackFailedToFetchMatchingAlerts(t *testing.T) {
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
		"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=key1%3D%22value1%22&filter=key2%3D%22value2%22&silenced=true&inhibited=true&active=true",
		httpmock.NewErrorResponder(errors.New("custom error")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")

	queryMatchers := types.QueryMatcherFromKeyValueString("key2=value2 key1=value1")
	key := app.Cache.Set(queryMatchers.GetHash(), queryMatchers.ToQueryString())

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Could not fetch alerts matching this silence: Get \"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=key1%3D%22value1%22&filter=key2%3D%22value2%22&silenced=true&inhibited=true&active=true\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

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
			Data:   key,
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
		"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=key1%3D%22value1%22&filter=key2%3D%22value2%22&silenced=true&inhibited=true&active=true",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-alerts.json")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")

	queryMatchers := types.QueryMatcherFromKeyValueString("key2=value2 key1=value1")
	key := app.Cache.Set(queryMatchers.GetHash(), queryMatchers.ToQueryString())

	queryMatchersWithoutFirst := queryMatchers.WithoutKey("key1")
	queryMatchersWithoutSecond := queryMatchers.WithoutKey("key2")

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytesAndMarkup(assets.GetBytesOrPanic("responses/silence-prepare-ok.html"), types.TelegramInlineKeyboardResponse{
			InlineKeyboard: [][]types.TelegramInlineKeyboard{
				{
					{
						Unique:       "grafana_prepare_silence_",
						Text:         "❌ Remove key1 = value1",
						CallbackData: fmt.Sprintf("\fgrafana_prepare_silence_|%s 1", queryMatchersWithoutFirst.GetHash()),
					},
				},
				{
					{
						Unique:       "grafana_prepare_silence_",
						Text:         "❌ Remove key2 = value2",
						CallbackData: fmt.Sprintf("\fgrafana_prepare_silence_|%s 1", queryMatchersWithoutSecond.GetHash()),
					},
				},
				{
					{
						Unique:       "grafana_silence_",
						Text:         "⌛ Silence for 1h",
						CallbackData: fmt.Sprintf("\fgrafana_silence_|%s 1h", key),
					},
				},
				{
					{
						Unique:       "grafana_silence_",
						Text:         "⌛ Silence for 3h",
						CallbackData: fmt.Sprintf("\fgrafana_silence_|%s 3h", key),
					},
				},
			},
		}),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

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
			Data:   key,
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
func TestAppPrepareSilenceViaCallbackOkWithEditKeyboard(t *testing.T) {
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
		"https://example.com/api/alertmanager/grafana/api/v2/alerts?filter=key1%3D%22value1%22&filter=key2%3D%22value2%22&silenced=true&inhibited=true&active=true",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-alerts.json")))

	app := NewApp(config, &fs.TestFS{}, "1.2.3")

	queryMatchers := types.QueryMatcherFromKeyValueString("key2=value2 key1=value1")
	key := app.Cache.Set(queryMatchers.GetHash(), queryMatchers.ToQueryString())

	queryMatchersWithoutFirst := queryMatchers.WithoutKey("key1")
	queryMatchersWithoutSecond := queryMatchers.WithoutKey("key2")

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		types.TelegramResponseHasBytesAndMarkup(assets.GetBytesOrPanic("responses/silence-prepare-ok.html"), types.TelegramInlineKeyboardResponse{
			InlineKeyboard: [][]types.TelegramInlineKeyboard{
				{
					{
						Unique:       "grafana_prepare_silence_",
						Text:         "❌ Remove key1 = value1",
						CallbackData: fmt.Sprintf("\fgrafana_prepare_silence_|%s 1", queryMatchersWithoutFirst.GetHash()),
					},
				},
				{
					{
						Unique:       "grafana_prepare_silence_",
						Text:         "❌ Remove key2 = value2",
						CallbackData: fmt.Sprintf("\fgrafana_prepare_silence_|%s 1", queryMatchersWithoutSecond.GetHash()),
					},
				},
				{
					{
						Unique:       "grafana_silence_",
						Text:         "⌛ Silence for 1h",
						CallbackData: fmt.Sprintf("\fgrafana_silence_|%s 1h", key),
					},
				},
				{
					{
						Unique:       "grafana_silence_",
						Text:         "⌛ Silence for 3h",
						CallbackData: fmt.Sprintf("\fgrafana_silence_|%s 3h", key),
					},
				},
			},
		}),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

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
			Data:   key + " 1",
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
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
			Data:   "123 48h",
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

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	queryMatchers := types.QueryMatcherFromKeyValueString("key2=value2 key1=value1")
	key := app.Cache.Set(queryMatchers.GetHash(), queryMatchers.ToQueryString())

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
			Data:   key + " 48h",
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
