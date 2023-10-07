package utils

import (
	"fmt"
	"main/pkg/logger"
	"main/pkg/types"
	"main/pkg/utils/normalize"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/exp/slices"

	tele "gopkg.in/telebot.v3"
)

func FindAlertRuleByName(groups []types.GrafanaAlertGroup, name string) (*types.GrafanaAlertRule, bool) {
	normalizedName := normalize.NormalizeString(name)

	for _, group := range groups {
		for _, rule := range group.Rules {
			ruleName := normalize.NormalizeString(group.Name + rule.Name)
			if strings.Contains(ruleName, normalizedName) {
				return &rule, true
			}
		}
	}

	return nil, false
}

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

func MergeMaps(first, second map[string]string) map[string]string {
	result := map[string]string{}

	for key, value := range first {
		result[key] = value
	}

	for key, value := range second {
		result[key] = value
	}

	return result
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

	matchers := ParseKeyValueString(rest)

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

func ParseKeyValueString(source string) []types.QueryMatcher {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)
		}
	}

	// splitting string by space but considering quoted section
	items := strings.FieldsFunc(source, f)

	matchers := make([]types.QueryMatcher, 0)
	operators := []string{"!=", "!~", "=~", "="}
	for index, item := range items {
		operatorFound := false
		for _, operator := range operators {
			if strings.Contains(item, operator) {
				operatorFound = true
				itemSplit := strings.Split(item, operator)
				matchers = append(matchers, types.QueryMatcher{
					Key:      itemSplit[0],
					Operator: operator,
					Value:    MaybeRemoveQuotes(itemSplit[1]),
				})
			}
		}

		if !operatorFound {
			matchers = append(matchers, types.QueryMatcher{
				Key:      "alertname",
				Operator: "=",
				Value:    strings.Join(items[index:], " "),
			})
			return matchers
		}
	}

	return matchers
}

func MaybeRemoveQuotes(source string) string {
	if len(source) > 0 && source[0] == '"' {
		source = source[1:]
	}
	if len(source) > 0 && source[len(source)-1] == '"' {
		source = source[:len(source)-1]
	}

	return source
}
