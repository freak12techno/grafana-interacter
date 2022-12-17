package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
	"strings"
	"time"
)

type App struct {
	Config       Config
	Grafana      *Grafana
	Alertmanager *Alertmanager
	Logger       *zerolog.Logger
	Bot          *tele.Bot
}

func NewApp(config *Config) *App {
	logger := GetLogger(config.Log)
	grafana := InitGrafana(config.Grafana, logger)
	alertmanager := InitAlertmanager(config.Alertmanager, logger)

	bot, err := tele.NewBot(tele.Settings{
		Token:   config.Telegram.Token,
		Poller:  &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: HandleError,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start Telegram bot")
	}

	if len(config.Telegram.Admins) > 0 {
		log.Debug().Msg("Using admins whitelist")
		bot.Use(middleware.Whitelist(config.Telegram.Admins...))
	}

	return &App{
		Logger:       logger,
		Grafana:      grafana,
		Alertmanager: alertmanager,
		Bot:          bot,
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
