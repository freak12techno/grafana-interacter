package app

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"strings"
	"sync"
	"testing"

	tele "gopkg.in/telebot.v3"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // disabled
func TestAppCannotFetchBot(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	config := &configPkg.Config{
		Timezone:     "Etc/GMT",
		Log:          configPkg.LogConfig{LogLevel: "info"},
		Telegram:     configPkg.TelegramConfig{Token: "xxx:yyy", Admins: []int64{1, 2}},
		Grafana:      configPkg.GrafanaConfig{URL: "https://example.com", User: "admin", Password: "admin"},
		Alertmanager: nil,
		Prometheus:   nil,
	}

	app := NewApp(config, "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppStartsOkay(t *testing.T) {
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

	app := NewApp(config, "1.2.3")
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		app.Start()

		ctx := app.Bot.NewContext(tele.Update{
			ID: 1,
			Message: &tele.Message{
				Sender: &tele.User{Username: "testuser"},
				Text:   "/help",
				Chat:   &tele.Chat{ID: 2},
			},
		})
		app.Bot.OnError(errors.New("custom error"), ctx)
		wg.Done()
	}()

	app.Stop()
	wg.Wait()
}

//nolint:paralleltest // disabled
func TestAppBotSendMultilineFail(t *testing.T) {
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

	err := app.BotReply(ctx, strings.Repeat("a", 5000))
	require.Error(t, err)

	err = app.BotReply(ctx, strings.Repeat("a", 10))
	require.Error(t, err)
}

//nolint:paralleltest // disabled
func TestAppBotSendMultilineOk(t *testing.T) {
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
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")))

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.BotReply(ctx, strings.Repeat("a", 5000))
	require.NoError(t, err)
}
