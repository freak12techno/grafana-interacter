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
func TestAppDeleteSilenceSilenceManagerDisabled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(false),
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
		types.TelegramResponseHasText("Grafana is disabled."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence 123",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleDeleteSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceInvalidInvocation(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		types.TelegramResponseHasText("Usage: /grafana_unsilence <silence ID or labels>"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleDeleteSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceFetchSilencesFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error getting silence to delete: Get \"https://example.com/api/alertmanager/grafana/api/v2/silences\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence 123",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleDeleteSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceSilenceNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Silence is not found by ID or matchers: 123"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence 123",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleDeleteSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceSilenceExpired(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Silence is already deleted!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence af780078-c86b-4c0d-bfbb-3edd72922f6d",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleDeleteSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceErrorDeletingSilence(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/4de5faa2-8c0c-4c66-bd31-25c3bf5fa231",
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error deleting silence: Delete \"https://example.com/api/alertmanager/grafana/api/v2/silence/4de5faa2-8c0c-4c66-bd31-25c3bf5fa231\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence 4de5faa2-8c0c-4c66-bd31-25c3bf5fa231",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleDeleteSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/api/alertmanager/grafana/api/v2/silence/4de5faa2-8c0c-4c66-bd31-25c3bf5fa231",
		httpmock.NewBytesResponder(200, []byte("")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/silence-delete-ok.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence 4de5faa2-8c0c-4c66-bd31-25c3bf5fa231",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleDeleteSilenceViaCommand(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceCallbackOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Silence is not found by ID or matchers: 123"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence 123",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaUnsilencePrefix,
			Data:   "123",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_unsilence 123",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleCallbackDeleteSilence(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppDeleteSilenceCallbackOkWithDeleteKeyboard(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.Config{
		Timezone: "Etc/GMT",
		Log:      configPkg.LogConfig{LogLevel: "info"},
		Telegram: configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana: configPkg.GrafanaConfig{
			URL:      "https://example.com",
			Silences: null.BoolFrom(true),
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
		"https://example.com/api/alertmanager/grafana/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Silence is not found by ID or matchers: 123"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, &fs.TestFS{}, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/grafana_unsilence 123",
			Chat:   &tele.Chat{ID: 2},
		},
		Callback: &tele.Callback{
			Sender: &tele.User{Username: "testuser"},
			Unique: "\f" + constants.GrafanaUnsilencePrefix,
			Data:   "123 1",
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/grafana_unsilence 123 1",
				Chat:   &tele.Chat{ID: 2},
			},
		},
	})

	err := app.HandleCallbackDeleteSilence(app.AlertSourcesWithSilenceManager[0].SilenceManager)(ctx)
	require.NoError(t, err)
}
