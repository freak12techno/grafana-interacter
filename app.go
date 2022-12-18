package main

import (
	"strings"
	"time"

	"github.com/rs/zerolog"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type App struct {
	Config          Config
	Grafana         *Grafana
	Alertmanager    *Alertmanager
	TemplateManager *TemplateManager
	Logger          *zerolog.Logger
	Bot             *tele.Bot
}

func NewApp(config *Config) *App {
	logger := GetLogger(config.Log)
	grafana := InitGrafana(config.Grafana, logger)
	alertmanager := InitAlertmanager(config.Alertmanager, logger)
	templateManager := NewTemplateManager()

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
	}
}

func (a *App) Start() {
	a.Bot.Handle("/dashboards", a.HandleListDashboards)
	a.Bot.Handle("/dashboard", a.HandleShowDashboard)
	a.Bot.Handle("/render", a.HandleRenderPanel)
	a.Bot.Handle("/datasources", a.HandleListDatasources)
	a.Bot.Handle("/alerts", a.HandleListAlerts)
	a.Bot.Handle("/firing", a.HandleListFiringAlerts)
	a.Bot.Handle("/alert", a.HandleSingleAlert)
	a.Bot.Handle("/silences", a.HandleListSilences)
	a.Bot.Handle("/silence", a.HandleNewSilence)
	a.Bot.Handle("/unsilence", a.HandleDeleteSilence)
	a.Bot.Handle("/alertmanager_silences", a.HandleAlertmanagerListSilences)
	a.Bot.Handle("/alertmanager_silence", a.HandleAlertmanagerNewSilence)
	a.Bot.Handle("/alertmanager_unsilence", a.HandleAlertmanagerDeleteSilence)

	a.Logger.Info().Msg("Telegram bot listening")

	a.Bot.Start()
}

func (a *App) BotReply(c tele.Context, msg string) error {
	msgsByNewline := strings.Split(msg, "\n")

	var sb strings.Builder

	for _, line := range msgsByNewline {
		if sb.Len()+len(line) > MaxMessageSize {
			if err := c.Reply(sb.String(), tele.ModeHTML); err != nil {
				a.Logger.Error().Err(err).Msg("Could not send Telegram message")
				return err
			}

			sb.Reset()
		}

		sb.WriteString(line + "\n")
	}

	if err := c.Reply(sb.String(), tele.ModeHTML); err != nil {
		a.Logger.Error().Err(err).Msg("Could not send Telegram message")
		return err
	}

	return nil
}
