package types

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/utils/generic"
	"net/url"
	"strings"
	"time"
)

type Silences []Silence

func (s Silences) FindByNameOrMatchers(source string) (*Silence, bool) {
	// Finding by ID first
	if silence, found := generic.Find(s, func(s Silence) bool {
		return s.ID == source
	}); found {
		return silence, true
	}

	queryMatchers := QueryMatcherFromKeyValueString(source)
	silenceMatchers := make(SilenceMatchers, len(queryMatchers))

	for index, queryMatcher := range queryMatchers {
		silenceMatcher := MatcherFromQueryMatcher(queryMatcher)
		silenceMatchers[index] = silenceMatcher
	}

	silenceFound, found := generic.Find(s, func(s Silence) bool {
		return s.Matchers.Equals(silenceMatchers)
	})

	return silenceFound, found
}

type Silence struct {
	Comment   string          `json:"comment"`
	CreatedBy string          `json:"createdBy"`
	StartsAt  time.Time       `json:"startsAt"`
	EndsAt    time.Time       `json:"endsAt"`
	ID        string          `json:"id,omitempty"`
	Matchers  SilenceMatchers `json:"matchers"`
	Status    SilenceStatus   `json:"status,omitempty"`
}

type SilenceMatchers []*SilenceMatcher

type SilenceMatcher struct {
	IsEqual bool   `json:"isEqual"`
	IsRegex bool   `json:"isRegex"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

type SilenceStatus struct {
	State string `json:"state"`
}

func (matcher *SilenceMatcher) Serialize() string {
	return fmt.Sprintf("%s %s %s", matcher.Name, matcher.GetSymbol(), matcher.Value)
}

func (matcher *SilenceMatcher) SerializeQueryString() string {
	return fmt.Sprintf("%s%s\"%s\"", matcher.Name, matcher.GetSymbol(), matcher.Value)
}

func (matcher *SilenceMatcher) GetSymbol() string {
	if matcher.IsEqual && matcher.IsRegex {
		return constants.SilenceMatcherRegexEqual
	} else if matcher.IsEqual && !matcher.IsRegex {
		return constants.SilenceMatcherEqual
	} else if !matcher.IsEqual && matcher.IsRegex {
		return constants.SilenceMatcherRegexNotEqual
	} else {
		return constants.SilenceMatcherNotEqual
	}
}

func (matcher *SilenceMatcher) Equals(otherMatcher *SilenceMatcher) bool {
	return matcher.IsEqual == otherMatcher.IsEqual &&
		matcher.IsRegex == otherMatcher.IsRegex &&
		matcher.Name == otherMatcher.Name &&
		matcher.Value == otherMatcher.Value
}

func (matchers SilenceMatchers) Equals(otherMatchers SilenceMatchers) bool {
	if len(matchers) != len(otherMatchers) {
		return false
	}

	for _, matcher := range matchers {
		_, found := generic.Find(otherMatchers, func(m *SilenceMatcher) bool {
			return m.Equals(matcher)
		})

		if !found {
			return false
		}
	}

	return true
}

func (matchers SilenceMatchers) GetFilterQueryString() string {
	filtersParts := generic.Map(matchers, func(m *SilenceMatcher) string {
		return "filter=" + url.QueryEscape(m.SerializeQueryString())
	})

	return strings.Join(filtersParts, "&")
}

type SilenceWithAlerts struct {
	Silence       Silence
	AlertsPresent bool
	Alerts        []AlertmanagerAlert
}
