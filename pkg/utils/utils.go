package utils

import (
	"fmt"
	"main/pkg/logger"
	"main/pkg/types"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
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
	keys := make([]string, len(qs))
	index := 0

	for key := range qs {
		keys[index] = key
		index++
	}

	sort.Strings(keys)

	tmp := make([]string, len(qs))
	counter := 0

	for _, key := range keys {
		tmp[counter] = key + "=" + qs[key]
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

func ParseSilenceFromCommand(query string, sender string) (*types.Silence, string) {
	args := strings.SplitN(query, " ", 3)
	if len(args) <= 2 {
		return nil, fmt.Sprintf("Usage: %s <duration> <params>", args[0])
	}

	cmd, args := args[0], args[1:] // removing first argument as it's always /silence
	durationString, rest := args[0], args[1]

	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return nil, "Invalid duration provided!"
	}

	matchers := types.QueryMatcherFromKeyValueString(rest)
	return ParseSilenceWithDuration(cmd, matchers, sender, duration)
}

func ParseSilenceWithDuration(
	cmd string,
	matchers types.QueryMatchers,
	sender string,
	duration time.Duration,
) (*types.Silence, string) {
	silence := &types.Silence{
		StartsAt:  time.Now(),
		EndsAt:    time.Now().Add(duration),
		Matchers:  []types.SilenceMatcher{},
		CreatedBy: sender,
		Comment: fmt.Sprintf(
			"Muted using grafana-interacter for %s by %s",
			duration,
			sender,
		),
	}

	for _, matcher := range matchers {
		if matcher.Key == "comment" {
			silence.Comment = matcher.Value
			continue
		}

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
			return nil, "Got unexpected operator: " + matcher.Operator
		}

		silence.Matchers = append(silence.Matchers, matcherParsed)
	}

	if len(silence.Matchers) == 0 {
		return nil, fmt.Sprintf("Usage: %s <duration> <params>", cmd)
	}

	return silence, ""
}

func StrToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Str("value", s).Msg("Could not parse float")
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

func FormatDate(timezone *time.Location) func(date time.Time) string {
	return func(date time.Time) string {
		return date.In(timezone).Format(time.RFC1123)
	}
}
