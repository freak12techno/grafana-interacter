package app

import (
	"main/pkg/alert_source"
	"main/pkg/clients"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/silence_manager"
	"main/pkg/templates"
	"strings"
	"time"

	"github.com/rs/zerolog"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
	templatesList "main/templates"
)

const MaxMessageSize = 4096

type AlertSourceWithSilenceManager struct {
	AlertSource    alert_source.AlertSource
	SilenceManager silence_manager.SilenceManager
}

type App struct {
	Config          *configPkg.Config
	Grafana         *clients.Grafana
	TemplateManager *templates.TemplateManager
	Logger          *zerolog.Logger
	Bot             *tele.Bot
	Version         string

	AlertSourcesWithSilenceManager []AlertSourceWithSilenceManager

	StopChannel chan bool
}

func NewApp(config *configPkg.Config, version string) *App {
	timezone, _ := time.LoadLocation(config.Timezone)

	logger := loggerPkg.GetLogger(config.Log)
	grafana := clients.InitGrafana(config.Grafana, logger)
	templateManager := templates.NewTemplateManager(timezone, templatesList.Templates)

	bot, err := tele.NewBot(tele.Settings{
		Token:  config.Telegram.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tele.Context) {
			logger.Error().Err(err).Msg("Telebot error")
		},
	})
	if err != nil {
		logger.Panic().Err(err).Msg("Could not start Telegram bot")
	}

	if len(config.Telegram.Admins) > 0 {
		logger.Debug().Msg("Using admins whitelist")
		bot.Use(middleware.Whitelist(config.Telegram.Admins...))
	}

	alertSourcesWithSilenceManagers := []AlertSourceWithSilenceManager{
		// Built-in Grafana alerting and silences
		{
			AlertSource:    alert_source.InitGrafana(config.Grafana, logger),
			SilenceManager: silence_manager.InitGrafana(config.Grafana, logger),
		},
		// External Prometheus alerts source and Alertmanager silence manager
		{
			AlertSource:    alert_source.InitPrometheus(config.Prometheus, logger),
			SilenceManager: silence_manager.InitAlertmanager(config.Alertmanager, logger),
		},
	}

	return &App{
		Config:                         config,
		Logger:                         logger,
		Grafana:                        grafana,
		TemplateManager:                templateManager,
		AlertSourcesWithSilenceManager: alertSourcesWithSilenceManagers,
		Bot:                            bot,
		Version:                        version,
		StopChannel:                    make(chan bool),
	}
}

func (a *App) Start() {
	a.Bot.Handle("/start", a.HandleHelp)
	a.Bot.Handle("/help", a.HandleHelp)
	a.Bot.Handle("/dashboards", a.HandleListDashboards)
	a.Bot.Handle("/dashboard", a.HandleShowDashboard)
	a.Bot.Handle("/render", a.HandleRenderPanel)
	a.Bot.Handle("/datasources", a.HandleListDatasources)
	a.Bot.Handle("/alerts", a.HandleListAlerts)
	a.Bot.Handle("/firing", a.HandleChooseAlertSourceForListAlerts)
	a.Bot.Handle("/alert", a.HandleSingleAlert)

	// TODO: fix
	a.Bot.Handle("/silences", a.HandleListSilences(a.AlertSourcesWithSilenceManager[0].SilenceManager))
	a.Bot.Handle("/silence", a.HandleNewSilenceViaCommand(a.AlertSourcesWithSilenceManager[0].SilenceManager))
	a.Bot.Handle("/unsilence", a.HandleDeleteSilenceViaCommand(a.AlertSourcesWithSilenceManager[0].SilenceManager))
	a.Bot.Handle("/alertmanager_silences", a.HandleListSilences(a.AlertSourcesWithSilenceManager[1].SilenceManager))
	a.Bot.Handle("/alertmanager_silence", a.HandleNewSilenceViaCommand(a.AlertSourcesWithSilenceManager[1].SilenceManager))
	a.Bot.Handle("/alertmanager_unsilence", a.HandleDeleteSilenceViaCommand(a.AlertSourcesWithSilenceManager[1].SilenceManager))

	// Callbacks

	for _, alertSourceWithSilenceManager := range a.AlertSourcesWithSilenceManager {
		alertSource := alertSourceWithSilenceManager.AlertSource
		silenceManager := alertSourceWithSilenceManager.SilenceManager
		alertSourcePrefixes := alertSourceWithSilenceManager.AlertSource.Prefixes()

		a.Bot.Handle("\f"+alertSourcePrefixes.PaginatedFiringAlerts, a.HandleListFiringAlertsFromCallback(alertSource, silenceManager))

		a.Bot.Handle("\f"+silenceManager.GetPaginatedSilencesListPrefix(), a.HandleListSilencesFromCallback(silenceManager))
		a.Bot.Handle("\f"+silenceManager.GetUnsilencePrefix(), a.HandleCallbackDeleteSilence(silenceManager))
		a.Bot.Handle("\f"+silenceManager.GetPrepareSilencePrefix(), a.HandlePrepareNewSilenceFromCallback(silenceManager, alertSource))
		a.Bot.Handle("\f"+silenceManager.GetSilencePrefix(), a.HandleCallbackNewSilence(silenceManager, alertSource))
	}

	a.Logger.Info().Msg("Telegram bot listening")

	go a.Bot.Start()

	<-a.StopChannel
	a.Logger.Info().Msg("Shutting down...")
	a.Bot.Stop()
}

func (a *App) BotReply(c tele.Context, msg string, opts ...interface{}) error {
	msgsByNewline := strings.Split(msg, "\n")

	var sb strings.Builder

	opts = append(opts, tele.ModeHTML, tele.NoPreview)

	for _, line := range msgsByNewline {
		if sb.Len()+len(line) > MaxMessageSize {
			if err := c.Reply(sb.String(), opts...); err != nil {
				a.Logger.Error().Err(err).Msg("Could not send Telegram message")
				return err
			}

			sb.Reset()
		}

		sb.WriteString(line + "\n")
	}

	if err := c.Reply(sb.String(), opts...); err != nil {
		a.Logger.Error().Err(err).Msg("Could not send Telegram message")
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.StopChannel <- true
}
