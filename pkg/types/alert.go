package types

import (
	"golang.org/x/exp/slices"
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

func (g GrafanaAlertGroups) FilterFiringOrPendingAlertGroups() []GrafanaAlertGroup {
	var returnGroups GrafanaAlertGroups

	alertingStatuses := []string{"firing", "alerting", "pending"}

	for _, group := range g {
		rules := []GrafanaAlertRule{}
		hasAnyRules := false

		for _, rule := range group.Rules {
			if !slices.Contains(alertingStatuses, strings.ToLower(rule.State)) {
				continue
			}

			alerts := []GrafanaAlert{}
			hasAnyAlerts := false

			for _, alert := range rule.Alerts {
				if !slices.Contains(alertingStatuses, strings.ToLower(alert.State)) {
					continue
				}

				alerts = append(alerts, alert)
				hasAnyAlerts = true
			}

			if hasAnyAlerts {
				rules = append(rules, GrafanaAlertRule{
					State:  rule.State,
					Name:   rule.Name,
					Alerts: alerts,
				})
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

type AlertmanagerAlert struct {
	Labels map[string]string `json:"labels"`
}
