package app

import (
	alertmanagerPkg "main/pkg/alertmanager"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	grafanaPkg "main/pkg/grafana"
	loggerPkg "main/pkg/logger"
	"main/pkg/templates"
	"strings"
	"time"

	"github.com/rs/zerolog"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

const MaxMessageSize = 4096

type App struct {
	Config          configPkg.Config
	Grafana         *grafanaPkg.Grafana
	Alertmanager    *alertmanagerPkg.Alertmanager
	TemplateManager *templates.TemplateManager
	Logger          *zerolog.Logger
	Bot             *tele.Bot
	Version         string
}

func NewApp(config *configPkg.Config, version string) *App {
	timezone, _ := time.LoadLocation(config.Timezone)

	logger := loggerPkg.GetLogger(config.Log)
	grafana := grafanaPkg.InitGrafana(config.Grafana, logger)
	alertmanager := alertmanagerPkg.InitAlertmanager(config.Alertmanager, logger)
	templateManager := templates.NewTemplateManager(timezone)

	bot, err := tele.NewBot(tele.Settings{
		Token:  config.Telegram.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tele.Context) {
			logger.Error().Err(err).Msg("Telebot error")
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start Telegram bot")
	}

	if len(config.Telegram.Admins) > 0 {
		logger.Debug().Msg("Using admins whitelist")
		bot.Use(middleware.Whitelist(config.Telegram.Admins...))
	}

	return &App{
		Logger:          logger,
		Grafana:         grafana,
		Alertmanager:    alertmanager,
		TemplateManager: templateManager,
		Bot:             bot,
		Version:         version,
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
	a.Bot.Handle("/silences", a.HandleGrafanaListSilences)
	a.Bot.Handle("/silence", a.HandleGrafanaNewSilence)
	a.Bot.Handle("/unsilence", a.HandleGrafanaDeleteSilence)
	a.Bot.Handle("/alertmanager_silences", a.HandleAlertmanagerListSilences)
	a.Bot.Handle("/alertmanager_silence", a.HandleAlertmanagerNewSilence)
	a.Bot.Handle("/alertmanager_unsilence", a.HandleAlertmanagerDeleteSilence)

	// Callbacks
	a.Bot.Handle("\f"+constants.GrafanaUnsilencePrefix, a.HandleGrafanaCallbackDeleteSilence)
	a.Bot.Handle("\f"+constants.AlertmanagerUnsilencePrefix, a.HandleAlertmanagerCallbackDeleteSilence)

	a.Logger.Info().Msg("Telegram bot listening")

	a.Bot.Start()
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
