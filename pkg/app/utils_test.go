package app

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/types/render"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestAppReplyRenderFailedToRender(t *testing.T) {
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
		types.TelegramResponseHasText("Error rendering template: template: pattern matches no files: `not_found.html`"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.ReplyRender(ctx, "not_found", render.RenderStruct{})
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppReplyRenderFailedToSend(t *testing.T) {
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
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		httpmock.NewErrorResponder(errors.New("custom error")))

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.ReplyRender(ctx, "not_found", render.RenderStruct{})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled
func TestAppEditRenderFailedToRender(t *testing.T) {
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
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		httpmock.NewErrorResponder(errors.New("custom error")))

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.EditRender(ctx, "not_found", render.RenderStruct{})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled
func TestAppEditRenderFailedToSend(t *testing.T) {
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
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageText",
		httpmock.NewErrorResponder(errors.New("custom error")))

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
				ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
					{
						{
							Text:   "text",
							Unique: "\f" + constants.GrafanaSilencePrefix,
							Data:   constants.GrafanaSilencePrefix + "|48h 123",
						},
					},
				}},
			},
		},
	})

	err := app.EditRender(ctx, "help", render.RenderStruct{})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled
func TestAppRemoveKeyboardItemFailedToDelete(t *testing.T) {
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
		"https://api.telegram.org/botxxx:yyy/editMessageReplyMarkup",
		types.TelegramResponseHasTextAndMarkup(
			"",
			types.TelegramInlineKeyboardResponse{
				InlineKeyboard: [][]types.TelegramInlineKeyboard{{
					{
						Text:         "text",
						Unique:       constants.GrafanaUnsilencePrefix,
						CallbackData: "\f" + constants.GrafanaUnsilencePrefix + "|random",
					},
				}},
			},
		),
		httpmock.NewErrorResponder(errors.New("custom error")),
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
				ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
					{
						{
							Text:   "text",
							Unique: "\f" + constants.GrafanaSilencePrefix,
							Data:   constants.GrafanaSilencePrefix + "|48h 123",
						},
						{
							Text:   "text",
							Unique: constants.GrafanaUnsilencePrefix,
							Data:   "random",
						},
					},
				}},
			},
		},
	})

	app.RemoveKeyboardItemByCallback(ctx, ctx.Callback())
}

//nolint:paralleltest // disabled
func TestAppClearKeyboardFailedToDelete(t *testing.T) {
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
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageReplyMarkup",
		httpmock.NewErrorResponder(errors.New("custom error")),
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
				ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
					{
						{
							Text:   "text",
							Unique: "\f" + constants.GrafanaSilencePrefix,
							Data:   constants.GrafanaSilencePrefix + "|48h 123",
						},
						{
							Text:   "text",
							Unique: constants.GrafanaUnsilencePrefix,
							Data:   "random",
						},
					},
				}},
			},
		},
	})

	err := app.ClearKeyboard(ctx)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled
func TestAppClearKeyboardOk(t *testing.T) {
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
		"POST",
		"https://api.telegram.org/botxxx:yyy/editMessageReplyMarkup",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")))

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
				ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
					{
						{
							Text:   "text",
							Unique: "\f" + constants.GrafanaSilencePrefix,
							Data:   constants.GrafanaSilencePrefix + "|48h 123",
						},
						{
							Text:   "text",
							Unique: constants.GrafanaUnsilencePrefix,
							Data:   "random",
						},
					},
				}},
			},
		},
	})

	err := app.ClearKeyboard(ctx)
	require.NoError(t, err)
}
