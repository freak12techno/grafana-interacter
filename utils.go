package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

const MAX_MESSAGE_SIZE = 4096

func NormalizeString(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(input, ""))
}

func FindDashboardByName(dashboards []GrafanaDashboardInfo, name string) (*GrafanaDashboardInfo, bool) {
	normalizedName := NormalizeString(name)

	for _, dashboard := range dashboards {
		if strings.Contains(NormalizeString(dashboard.Title), normalizedName) {
			return &dashboard, true
		}
	}

	return nil, false
}

func FindPanelByName(panels []PanelStruct, name string) (*PanelStruct, bool) {
	normalizedName := NormalizeString(name)

	for _, panel := range panels {
		panelNameWithDashboardName := NormalizeString(panel.DashboardName + panel.Name)

		if strings.Contains(panelNameWithDashboardName, normalizedName) {
			return &panel, true
		}
	}

	return nil, false
}

func FindAlertRuleByName(groups []GrafanaAlertGroup, name string) (*GrafanaAlertRule, bool) {
	normalizedName := NormalizeString(name)

	for _, group := range groups {
		for _, rule := range group.Rules {
			ruleName := NormalizeString(group.Name + rule.Name)
			if strings.Contains(ruleName, normalizedName) {
				return &rule, true
			}
		}
	}

	return nil, false
}

func ParseRenderOptions(query string) (RenderOptions, bool) {
	args := strings.Split(query, " ")
	if len(args) <= 1 {
		return RenderOptions{}, false // should have at least 1 argument
	}

	params := map[string]string{}

	_, args = args[0], args[1:] // removing first argument as it's always /render
	for len(args) > 0 {
		if !strings.Contains(args[0], "=") {
			break
		}

		paramSplit := strings.SplitN(args[0], "=", 2)
		params[paramSplit[0]] = paramSplit[1]

		_, args = args[0], args[1:]
	}

	return RenderOptions{
		Query:  strings.Join(args, " "),
		Params: params,
	}, len(args) > 0
}

func SerializeQueryString(qs map[string]string) string {
	tmp := make([]string, len(qs))
	counter := 0

	for key, value := range qs {
		tmp[counter] = key + "=" + value
		counter++
	}

	return strings.Join(tmp, "&")
}

func MergeMaps(first, second map[string]string) map[string]string {
	for key, value := range second {
		first[key] = value
	}

	return first
}

func GetEmojiByStatus(state string) string {
	switch strings.ToLower(state) {
	case "inactive", "ok", "normal":
		return "????"
	case "pending":
		return "????"
	case "firing", "alerting":
		return "????"
	default:
		return "[" + state + "]"
	}
}

func GetEmojiBySilenceStatus(state string) string {
	switch strings.ToLower(state) {
	case "active":
		return "????"
	case "expired":
		return "???"
	default:
		return "[" + state + "]"
	}
}

func SerializeAlertLabels(qs map[string]string) string {
	tmp := make([]string, len(qs))
	counter := 0

	for key, value := range qs {
		if strings.HasPrefix(key, "__") && strings.HasSuffix(key, "__") {
			continue
		}
		tmp[counter] = key + "=" + value
		counter++
	}

	return strings.Join(tmp, " ")
}

func BotReply(c tele.Context, msg string) error {
	msgsByNewline := strings.Split(msg, "\n")

	var sb strings.Builder

	for _, line := range msgsByNewline {
		if sb.Len()+len(line) > MAX_MESSAGE_SIZE {
			if err := c.Reply(sb.String(), tele.ModeHTML); err != nil {
				log.Error().Err(err).Msg("Could not send Telegram message")
				return err
			}

			sb.Reset()
		}

		sb.WriteString(line + "\n")
	}

	if err := c.Reply(sb.String(), tele.ModeHTML); err != nil {
		log.Error().Err(err).Msg("Could not send Telegram message")
		return err
	}

	return nil
}

func ParseSilenceOptions(query string, c tele.Context) (*Silence, string) {
	args := strings.Split(query, " ")
	if len(args) <= 2 {
		return nil, "Usage: /silence <duration> <params>"
	}

	_, args = args[0], args[1:] // removing first argument as it's always /silence
	durationString, args := args[0], args[1:]

	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return nil, "Invalid duration provided"
	}

	silence := Silence{
		StartsAt:  time.Now(),
		EndsAt:    time.Now().Add(duration),
		Matchers:  []SilenceMatcher{},
		CreatedBy: c.Sender().FirstName,
		Comment: fmt.Sprintf(
			"Muted using grafana-interacter for %s by %s",
			duration,
			c.Sender().FirstName,
		),
	}

	for len(args) > 0 {
		if strings.Contains(args[0], "!=") {
			// not equals
			argsSplit := strings.SplitN(args[0], "!=", 2)
			silence.Matchers = append(silence.Matchers, SilenceMatcher{
				IsEqual: false,
				IsRegex: false,
				Name:    argsSplit[0],
				Value:   argsSplit[1],
			})
		} else if strings.Contains(args[0], "!~") {
			// not matches regexp
			argsSplit := strings.SplitN(args[0], "!~", 2)
			silence.Matchers = append(silence.Matchers, SilenceMatcher{
				IsEqual: false,
				IsRegex: true,
				Name:    argsSplit[0],
				Value:   argsSplit[1],
			})
		} else if strings.Contains(args[0], "=~") {
			// matches regexp
			argsSplit := strings.SplitN(args[0], "=~", 2)
			silence.Matchers = append(silence.Matchers, SilenceMatcher{
				IsEqual: true,
				IsRegex: true,
				Name:    argsSplit[0],
				Value:   argsSplit[1],
			})
		} else if strings.Contains(args[0], "=") {
			// equals
			argsSplit := strings.SplitN(args[0], "=", 2)
			silence.Matchers = append(silence.Matchers, SilenceMatcher{
				IsEqual: true,
				IsRegex: false,
				Name:    argsSplit[0],
				Value:   argsSplit[1],
			})
		} else {
			break
		}

		_, args = args[0], args[1:]
	}

	if len(args) > 0 {
		// plain string, silencing by alertname
		silence.Matchers = append(silence.Matchers, SilenceMatcher{
			IsEqual: true,
			IsRegex: false,
			Name:    "alertname",
			Value:   strings.Join(args, " "),
		})
	}

	if len(silence.Matchers) == 0 {
		return nil, "Usage: /silence <duration> <params>"
	}

	return &silence, ""
}

func FilterFiringOrPendingAlertGroups(groups []GrafanaAlertGroup) []GrafanaAlertGroup {
	returnGroups := []GrafanaAlertGroup{}

	for _, group := range groups {
		rules := []GrafanaAlertRule{}
		hasAnyRules := false

		for _, rule := range group.Rules {
			if rule.State == "firing" || rule.State == "alerting" || rule.State == "pending" {
				rules = append(rules, rule)
				hasAnyRules = true
			}
		}

		if hasAnyRules {
			returnGroups = append(returnGroups, GrafanaAlertGroup{
				Name:  group.Name,
				File:  group.File,
				Rules: rules,
			})
		}
	}

	return returnGroups
}
