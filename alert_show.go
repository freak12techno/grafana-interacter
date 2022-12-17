package main

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) HandleSingleAlert(c tele.Context) error {
	a.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got single alert query")

	args := strings.SplitN(c.Text(), " ", 2)
	_, args = args[0], args[1:] // removing first argument as it's always /alert

	if len(args) != 1 {
		return c.Reply("Usage: /alert <alert name>")
	}

	rules, err := a.Grafana.GetAllAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	rule, found := FindAlertRuleByName(rules, args[0])
	if !found {
		return c.Reply("Could not find alert. See /alert for alerting rules.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<strong>Alert rule: </strong>%s\n", rule.Name))
	sb.WriteString("<strong>Alerts: </strong>\n")

	for _, alert := range rule.Alerts {
		sb.WriteString(alert.Serialize())
	}

	return a.BotReply(c, sb.String())
}
