package utils

import (
	"fmt"
	"main/pkg/logger"
	"main/pkg/types"
	"math"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	tele "gopkg.in/telebot.v3"
)

func ParseRenderOptions(query string) (types.RenderOptions, bool) {
	args := strings.Split(query, " ")
	if len(args) <= 1 {
		return types.RenderOptions{}, false // should have at least 1 argument
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

	return types.RenderOptions{
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

func GetEmojiByStatus(state string) string {
	switch strings.ToLower(state) {
	case "inactive", "ok", "normal":
		return "🟢"
	case "pending":
		return "🟡"
	case "firing", "alerting":
		return "🔴"
	default:
		return "[" + state + "]"
	}
}

func GetEmojiBySilenceStatus(state string) string {
	switch strings.ToLower(state) {
	case "active":
		return "🟢"
	case "expired":
		return "⚪"
	default:
		return "[" + state + "]"
	}
}

func ParseSilenceOptions(query string, c tele.Context) (*types.Silence, string) {
	args := strings.SplitN(query, " ", 3)
	if len(args) <= 2 {
		return nil, fmt.Sprintf("Usage: %s <duration> <params>", args[0])
	}

	_, args = args[0], args[1:] // removing first argument as it's always /silence
	durationString, rest := args[0], args[1]

	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return nil, "Invalid duration provided"
	}

	silence := types.Silence{
		StartsAt:  time.Now(),
		EndsAt:    time.Now().Add(duration),
		Matchers:  []types.SilenceMatcher{},
		CreatedBy: c.Sender().FirstName,
		Comment: fmt.Sprintf(
			"Muted using grafana-interacter for %s by %s",
			duration,
			c.Sender().FirstName,
		),
	}

	matchers := types.QueryMatcherFromKeyValueString(rest)

	for _, matcher := range matchers {
		matcherParsed := types.SilenceMatcher{
			Name:  matcher.Key,
			Value: matcher.Value,
		}

		switch matcher.Operator {
		case "!=":
			matcherParsed.IsEqual = false
			matcherParsed.IsRegex = false
		case "!~":
			matcherParsed.IsEqual = false
			matcherParsed.IsRegex = true
		case "=~":
			matcherParsed.IsEqual = true
			matcherParsed.IsRegex = true
		case "=":
			matcherParsed.IsEqual = true
			matcherParsed.IsRegex = false
		default:
			return nil, fmt.Sprintf("Got unexpected operator: %s", matcher.Operator)
		}

		silence.Matchers = append(silence.Matchers, matcherParsed)
	}

	if len(silence.Matchers) == 0 {
		return nil, "Usage: /silence <duration> <params>"
	}

	return &silence, ""
}

func FilterFiringOrPendingAlertGroups(groups []types.GrafanaAlertGroup) []types.GrafanaAlertGroup {
	var returnGroups []types.GrafanaAlertGroup

	alertingStatuses := []string{"firing", "alerting", "pending"}

	for _, group := range groups {
		rules := []types.GrafanaAlertRule{}
		hasAnyRules := false

		for _, rule := range group.Rules {
			if !slices.Contains(alertingStatuses, strings.ToLower(rule.State)) {
				continue
			}

			alerts := []types.GrafanaAlert{}
			hasAnyAlerts := false

			for _, alert := range rule.Alerts {
				if !slices.Contains(alertingStatuses, strings.ToLower(alert.State)) {
					continue
				}

				alerts = append(alerts, alert)
				hasAnyAlerts = true
			}

			if hasAnyAlerts {
				rules = append(rules, types.GrafanaAlertRule{
					State:  rule.State,
					Name:   rule.Name,
					Alerts: alerts,
				})
				hasAnyRules = true
			}
		}

		if hasAnyRules {
			returnGroups = append(returnGroups, types.GrafanaAlertGroup{
				Name:  group.Name,
				File:  group.File,
				Rules: rules,
			})
		}
	}

	return returnGroups
}

func StrToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Str("value", s).Msg("Could not parse float")
	}

	return f
}

func FormatDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"day", days},
		{"hour", hours},
		{"minute", minutes},
		{"second", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		case 1:
			parts = append(parts, fmt.Sprintf("%d %s", chunk.amount, chunk.singularName))
		default:
			parts = append(parts, fmt.Sprintf("%d %ss", chunk.amount, chunk.singularName))
		}
	}

	return strings.Join(parts, " ")
}

func FormatDate(date time.Time) string {
	return date.Format(time.RFC1123)
}
