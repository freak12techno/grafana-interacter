package types

import (
	"fmt"
	"main/pkg/constants"
	"sort"
	"strings"
	"unicode"
)

type RenderOptions struct {
	Query  string
	Params map[string]string
}

type SilenceCreateResponse struct {
	SilenceID string `json:"silenceID"`
}

type QueryMatcher struct {
	Key      string
	Operator string
	Value    string
}

func (matcher *QueryMatcher) Serialize() string {
	return fmt.Sprintf("%s %s %s", matcher.Key, matcher.Operator, matcher.Value)
}

func QueryMatcherFromKeyValueString(source string) QueryMatchers {
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

	matchers := make([]QueryMatcher, 0)
	operators := []string{"!=", "!~", "=~", "="}
	for index, item := range items {
		operatorFound := false
		for _, operator := range operators {
			if strings.Contains(item, operator) {
				operatorFound = true
				itemSplit := strings.Split(item, operator)
				matchers = append(matchers, QueryMatcher{
					Key:      itemSplit[0],
					Operator: operator,
					Value:    MaybeRemoveQuotes(itemSplit[1]),
				})
				break
			}
		}

		if !operatorFound {
			matchers = append(matchers, QueryMatcher{
				Key:      "alertname",
				Operator: "=",
				Value:    strings.Join(items[index:], " "),
			})
			return matchers
		}
	}

	return matchers
}

type QueryMatchers []QueryMatcher

func (q QueryMatchers) Sort() {
	sort.Slice(q, func(i, j int) bool {
		return q[i].Key < q[j].Key
	})
}

func MatcherFromQueryMatcher(queryMatcher QueryMatcher) *SilenceMatcher {
	matcherParsed := &SilenceMatcher{
		Name:  queryMatcher.Key,
		Value: queryMatcher.Value,
	}

	switch queryMatcher.Operator {
	case constants.SilenceMatcherNotEqual:
		matcherParsed.IsEqual = false
		matcherParsed.IsRegex = false
	case constants.SilenceMatcherRegexNotEqual:
		matcherParsed.IsEqual = false
		matcherParsed.IsRegex = true
	case constants.SilenceMatcherRegexEqual:
		matcherParsed.IsEqual = true
		matcherParsed.IsRegex = true
	case constants.SilenceMatcherEqual:
		matcherParsed.IsEqual = true
		matcherParsed.IsRegex = false
	}

	return matcherParsed
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
