package types

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"main/pkg/utils/generic"
	"main/pkg/utils/normalize"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type GrafanaAlertRulesResponse struct {
	Data GrafanaAlertRulesData `json:"data"`
}

type GrafanaAlertRulesData struct {
	Groups GrafanaAlertGroups `json:"groups"`
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

func (g GrafanaAlertRule) SerializeAlertsCount() string {
	firing := generic.Filter(g.Alerts, func(a GrafanaAlert) bool {
		return a.State == "firing"
	})

	pending := generic.Filter(g.Alerts, func(a GrafanaAlert) bool {
		return a.State == "pending"
	})

	array := []string{}

	if len(firing) > 0 {
		array = append(array, fmt.Sprintf("%d firing", len(firing)))
	}

	if len(pending) > 0 {
		array = append(array, fmt.Sprintf("%d pending", len(pending)))
	}

	if len(array) == 0 {
		return ""
	}

	return " (" + strings.Join(array, ", ") + ")"
}

type GrafanaAlert struct {
	Labels   map[string]string `json:"labels"`
	State    string            `json:"state"`
	Value    string            `json:"value"`
	ActiveAt time.Time         `json:"activeAt"`
}

func (a GrafanaAlert) GetHash() string {
	hash := md5.Sum([]byte(a.SerializeLabels()))
	return hex.EncodeToString(hash[:])[0:8]
}

func (a GrafanaAlert) SerializeLabels() string {
	keys := make([]string, len(a.Labels))
	index := 0
	for key := range a.Labels {
		keys[index] = key
		index++
	}

	slices.Sort(keys)

	labels := make([]string, len(a.Labels))
	for keyIndex, key := range keys {
		labels[keyIndex] = fmt.Sprintf("%s=%s", key, a.Labels[key])
	}

	return strings.Join(labels, " ")
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

func (g GrafanaAlertGroups) FilterFiringOrPendingAlertGroups(leavePending bool) GrafanaAlertGroups {
	var returnGroups GrafanaAlertGroups

	alertingStatuses := []string{"firing", "alerting", "pending"}
	if !leavePending {
		alertingStatuses = []string{"firing", "alerting"}
	}

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

func (g GrafanaAlertGroups) ToFiringAlerts() []FiringAlert {
	firingAlerts := make([]FiringAlert, 0)

	for _, alertGroup := range g {
		for _, alertRule := range alertGroup.Rules {
			for _, alert := range alertRule.Alerts {
				firingAlerts = append(firingAlerts, FiringAlert{
					GroupName:     alertGroup.Name,
					Alert:         alert,
					AlertRuleName: alertRule.Name,
				})
			}
		}
	}

	return firingAlerts
}

type AlertmanagerAlert struct {
	Labels map[string]string `json:"labels"`
}
