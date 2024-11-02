package types

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"main/pkg/utils/normalize"
	"strings"
	"time"

	"golang.org/x/exp/slices"
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

func (a GrafanaAlert) GetCallbackHash() string {
	// Using hash here as Telegram limits callback size to 64 chars
	// Firstly, need to make sure it's ordered, then convert it to a string
	// like "label1=value1 label2=value2", then take a md5 hash of it.
	hash := md5.Sum([]byte(a.SerializeLabels()))
	return hex.EncodeToString(hash[:])
}

func (a GrafanaAlert) SerializeLabels() string {
	// Using hash here as Telegram limits callback size to 64 chars
	// Firstly, need to make sure it's ordered, then convert it to a string
	// like "label1=value1 label2=value2", then take a md5 hash of it.
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
