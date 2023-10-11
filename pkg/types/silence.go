package types

import (
	"fmt"
	"main/pkg/utils/generic"
	"time"
)

type Silences []Silence

func (s Silences) FindByNameOrMatchers(source string) (*Silence, bool, string) {
	// Finding by ID first
	if silence, found := generic.Find(s, func(s Silence) bool {
		return s.ID == source
	}); found {
		return silence, true, ""
	}

	queryMatchers := QueryMatcherFromKeyValueString(source)
	silenceMatchers := make(SilenceMatchers, len(queryMatchers))

	for index, queryMatcher := range queryMatchers {
		silenceMatcher, err := MatcherFromQueryMatcher(queryMatcher)
		if err != "" {
			return nil, false, err
		}
		silenceMatchers[index] = *silenceMatcher
	}

	silenceFound, found := generic.Find(s, func(s Silence) bool {
		return s.Matchers.Equals(silenceMatchers)
	})

	return silenceFound, found, ""
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

type SilenceMatchers []SilenceMatcher

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
	if matcher.IsEqual && matcher.IsRegex {
		return fmt.Sprintf("%s ~= %s", matcher.Name, matcher.Value)
	} else if matcher.IsEqual && !matcher.IsRegex {
		return fmt.Sprintf("%s = %s", matcher.Name, matcher.Value)
	} else if !matcher.IsEqual && matcher.IsRegex {
		return fmt.Sprintf("%s !~ %s", matcher.Name, matcher.Value)
	} else {
		return fmt.Sprintf("%s != %s", matcher.Name, matcher.Value)
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
		_, found := generic.Find(otherMatchers, func(m SilenceMatcher) bool {
			return m.Equals(&matcher)
		})

		if !found {
			return false
		}
	}

	return true
}
