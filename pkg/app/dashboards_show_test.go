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
func TestAppShowDashboardInvalidInvocation(t *testing.T) {
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
		types.TelegramResponseHasText("Usage: /dashboard <dashboard>"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/dashboard",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleShowDashboard(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppShowDashboardDashboardsListFail(t *testing.T) {
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
		types.TelegramResponseHasText("Error querying for dashboards: Get \"https://example.com/api/search?type=dash-db\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/dashboard id",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleShowDashboard(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppShowDashboardDashboardNotFound(t *testing.T) {
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

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Could not find dashboard. See /dashboards for dashboards list."),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/dashboard test",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleShowDashboard(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppShowDashboardDashboardFetchError(t *testing.T) {
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
		httpmock.NewErrorResponder(errors.New("custom error")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Could not get dashboard: Get \"https://example.com/api/dashboards/uid/alertmanager\": custom error"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/dashboard alertmanager",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleShowDashboard(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestAppShowDashboardOk(t *testing.T) {
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
		types.TelegramResponseHasText("\n<strong>Dashboard <a href='https://example.com/d/alertmanager/alertmanager'>Alertmanager</a></strong>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=36'>General info</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=4'>Number of instances</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=26'>Instance versions and up time</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=207'>Cluster size</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=2'>Number of active alerts</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=3'>Number of suppressed alerts</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=121'>Number of active silences</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=113'>Notifications</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=118'>Notifications sent from $instance</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=115'>Notification durations per integration on $instance</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=18'>Alerts</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=6'>Active alerts in $instance</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=8'>Received alerts by status for $instance</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=34'>Cluster members</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=284'>Gossip messages</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=84'>Nflog</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=123'>Silences</a>\n- <a href='https://example.com/d/alertmanager/alertmanager?viewPanel=173'>Resources</a>\n"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	app := NewApp(config, "1.2.3")
	ctx := app.Bot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/dashboard alertmanager",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := app.HandleShowDashboard(ctx)
	require.NoError(t, err)
}
