package app

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/types/render"
	"strings"
	"time"

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

	rules, err := a.GetAllAlertingRules()
	if err != nil {
		return c.Reply(fmt.Sprintf("Error querying alerts: %s", err))
	}

	rule, found := rules.FindAlertRuleByName(args[0])
	if !found {
		return c.Reply("Could not find alert. See /alerts for alerting rules.")
	}

	return a.ReplyRender(c, "alert", render.RenderStruct{
		Grafana: a.Grafana,
		Data: types.SingleAlertStruct{
			Alert:      rule,
			RenderTime: time.Now(),
		},
	})
}
