package app

import (
	"main/pkg/alert_source"
	"main/pkg/clients"
	configPkg "main/pkg/config"
	"main/pkg/constants"
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

type App struct {
	Config          *configPkg.Config
	Grafana         *clients.Grafana
	Alertmanager    *silence_manager.Alertmanager
	Prometheus      *alert_source.Prometheus
	TemplateManager *templates.TemplateManager
	Logger          *zerolog.Logger
	Bot             *tele.Bot
	Version         string

	StopChannel chan bool
}

func NewApp(config *configPkg.Config, version string) *App {
	timezone, _ := time.LoadLocation(config.Timezone)

	logger := loggerPkg.GetLogger(config.Log)
	grafana := clients.InitGrafana(config.Grafana, logger)
	alertmanager := silence_manager.InitAlertmanager(config.Alertmanager, logger)
	prometheus := alert_source.InitPrometheus(config.Prometheus, logger)
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

	return &App{
		Config:          config,
		Logger:          logger,
		Grafana:         grafana,
		Alertmanager:    alertmanager,
		Prometheus:      prometheus,
		TemplateManager: templateManager,
		Bot:             bot,
		Version:         version,
		StopChannel:     make(chan bool),
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
	a.Bot.Handle("/firing", a.HandleListFiringAlerts)
	a.Bot.Handle("/alert", a.HandleSingleAlert)
	a.Bot.Handle("/silences", a.HandleListSilences(a.Grafana))
	a.Bot.Handle("/silence", a.HandleNewSilenceViaCommand(a.Grafana))
	a.Bot.Handle("/unsilence", a.HandleDeleteSilenceViaCommand(a.Grafana))
	a.Bot.Handle("/alertmanager_silences", a.HandleListSilences(a.Alertmanager))
	a.Bot.Handle("/alertmanager_silence", a.HandleNewSilenceViaCommand(a.Alertmanager))
	a.Bot.Handle("/alertmanager_unsilence", a.HandleDeleteSilenceViaCommand(a.Alertmanager))

	// Callbacks
	a.Bot.Handle("\f"+constants.GrafanaPaginatedSilencesList, a.HandleListSilencesFromCallback(a.Grafana))
	a.Bot.Handle("\f"+constants.AlertmanagerPaginatedSilencesList, a.HandleListSilencesFromCallback(a.Alertmanager))
	a.Bot.Handle("\f"+constants.GrafanaUnsilencePrefix, a.HandleCallbackDeleteSilence(a.Grafana))
	a.Bot.Handle("\f"+constants.AlertmanagerUnsilencePrefix, a.HandleCallbackDeleteSilence(a.Alertmanager))
	a.Bot.Handle("\f"+constants.GrafanaPrepareSilencePrefix, a.HandlePrepareNewSilenceFromCallback(a.Grafana, a.Grafana))
	a.Bot.Handle("\f"+constants.AlertmanagerPrepareSilencePrefix, a.HandlePrepareNewSilenceFromCallback(a.Alertmanager, a.Prometheus))
	a.Bot.Handle("\f"+constants.GrafanaSilencePrefix, a.HandleCallbackNewSilence(a.Grafana, a.Grafana))
	a.Bot.Handle("\f"+constants.AlertmanagerSilencePrefix, a.HandleCallbackNewSilence(a.Alertmanager, a.Prometheus))

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
