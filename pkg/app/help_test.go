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
func TestAppHelpFailedToRender(t *testing.T) {
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
		types.TelegramResponseHasText("Error rendering template: template: pattern matches no files: `help.html`"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	app.TemplateManager.Filesystem = assets.EmbedFS
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleHelp(ctx)
	require.NoError(t, err) // we did send a message about the failed render
}

//nolint:paralleltest // disabled
func TestAppHelpFailedToSend(t *testing.T) {
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

	err := app.HandleHelp(ctx)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled
func TestAppHelpOk(t *testing.T) {
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
		types.TelegramResponseHasText("<a href=\"https://github.com/freak12techno/grafana-interacter\">grafana-interacter</a> v1.2.3\nA Telegram bot that allows you to interact with your Grafana, Prometheus and Alertmanager instances.\nCan understand the following commands:\n\n- /help, or /start - displays this message\n- /render [opts] panelname - renders the panel and sends it as image. If there are multiple panels with the same name (for example, you have a 'dashboard1' and 'dashboard2' both containing panel with name 'panel'), it will render the first panel it will find. For specifying it, you may add the dashboard name as a prefix to your query (like <code>/render dashboard1 panel</code>). You can also provide options in a 'key=value' format, which will be internally passed to a <code>/render</code> query to Grafana. Some examples are 'from', 'to', 'width', 'height' (the command would look something like <code>/render from=now-14d to=now-7d width=100 height=100 dashboard1 panel</code>). By default, the params are: <code>width=1000&height=500&from=now-30m&to=now&tz=Europe/Moscow</code>.\n- /dashboards - will list Grafana dashboards and links to them.\n- /dashboard [name] - will return a link to a dashboard and its panels.\n- /datasources - will return Grafana datasources.\n- /alerts - will list both Grafana alerts and Prometheus alerts from all Prometheus datasources, if any\n- /firing - will list firing and pending alerts from both Grafana and Prometheus datasources, along with their details\n- /silence [duration] [params] - creates a silence for Grafana alert. You need to pass a duration (like <code>/silence 2h test alert</code>) and some params for matching alerts to silence. You may use '=' for matching the value exactly (example: <code>/silence 2h host=localhost</code>), '!=' for matching everything except this value (example:  <code>/silence 2h host!=localhost</code>), '=~' for matching everything that matches the regexp (example:  <code>/silence 2h host=~local</code>), '!~' for matching everything that doesn't match the regexp (example:  <code>/silence 2h host!~local</code>), or just provide a string that will be treated as an alert name (example:  <code>/silence 2h test alert</code>).\n- /silences - list silences (both active and expired).\n- /unsilence [silence ID or labels] - deletes a silence. You can pass either a silence ID (like <code>/unsilence xxxx</code>, or labels set (like <code>/unsilence host=test</code>) as an argument.\n- /alertmanager_silences - same as /silences, but using external Alertmanager.\n- /alertmanager_silence - same as /silence, but using external Alertmanager.\n- /alertmanager_unsilence - same as /unsilence, but using external Alertmanager.\n\nCreated by <a href=\"https://github.com/freak12techno\">freak12techno</a> with ❤️.\n\n"),
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

	err := app.HandleHelp(ctx)
	require.NoError(t, err)
}
