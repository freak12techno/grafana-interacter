package types

import (
	"main/pkg/utils/normalize"
	"strings"
	"time"
)

type GrafanaAlertRulesResponse struct {
	Data GrafanaAlertRulesData `json:"data"`
}

type GrafanaAlertRulesData struct {
	Groups []GrafanaAlertGroup `json:"groups"`
}

type GrafanaAlertGroup struct {
	Name  string             `json:"name"`
	File  string             `json:"file"`
	Rules []GrafanaAlertRule `json:"rules"`
}

type GrafanaAlertRule struct {
	State  string         `json:"state"`
	Name   string         `json:"name"`
	Alerts []GrafanaAlert `json:"alerts"`
}

type GrafanaAlert struct {
	Labels   map[string]string `json:"labels"`
	State    string            `json:"state"`
	Value    string            `json:"value"`
	ActiveAt time.Time         `json:"activeAt"`
}

func (a GrafanaAlert) ActiveSince() time.Duration {
	return time.Since(a.ActiveAt)
}

type GrafanaAlertGroups []GrafanaAlertGroup

func (g GrafanaAlertGroups) FindAlertRuleByName(name string) (*GrafanaAlertRule, bool) {
	normalizedName := normalize.NormalizeString(name)

	for _, group := range g {
		for _, rule := range group.Rules {
			ruleName := normalize.NormalizeString(group.Name + rule.Name)
			if strings.Contains(ruleName, normalizedName) {
				return &rule, true
			}
		}
	}

	return nil, false
}
