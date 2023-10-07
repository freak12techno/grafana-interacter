package types

import (
	"fmt"
	"time"
)

type Silence struct {
	Comment   string           `json:"comment"`
	CreatedBy string           `json:"createdBy"`
	StartsAt  time.Time        `json:"startsAt"`
	EndsAt    time.Time        `json:"endsAt"`
	ID        string           `json:"id,omitempty"`
	Matchers  []SilenceMatcher `json:"matchers"`
	Status    SilenceStatus    `json:"status,omitempty"`
}

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
