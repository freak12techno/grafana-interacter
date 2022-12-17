package main

import (
	"fmt"
	"strings"
	"time"
)

type PanelStruct struct {
	Name          string
	DashboardName string
	DashboardID   string
	DashboardURL  string
	PanelID       int
}

type GrafanaDashboardInfo struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type GrafanaDashboardResponse struct {
	Dashboard GrafanaSingleDashboard `json:"dashboard"`
	Meta      GrafanaDashboardMeta   `json:"meta"`
}

type GrafanaSingleDashboard struct {
	Title  string         `json:"title"`
	UID    string         `json:"uid"`
	Panels []GrafanaPanel `json:"panels"`
}

type GrafanaDashboardMeta struct {
	URL string `json:"url"`
}

type GrafanaPanel struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type GrafanaDatasource struct {
	ID   int    `json:"id"`
	UID  string `json:"uid"`
	Name string `json:"name"`
	Type string `json:"type"`
}

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
	Labels map[string]string `json:"labels"`
	State  string            `json:"state"`
	Value  string            `json:"value"`
}

type RenderOptions struct {
	Query  string
	Params map[string]string
}

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

func (rule *GrafanaAlertRule) Serialize(groupName string) string {
	return fmt.Sprintf("- %s %s -> %s\n", GetEmojiByStatus(rule.State), groupName, rule.Name)
}

func (alert *GrafanaAlert) Serialize() string {
	return fmt.Sprintf("- %s <pre>%s</pre>", GetEmojiByStatus(alert.State), SerializeAlertLabels(alert.Labels))
}

func (rule *GrafanaAlertRule) SerializeFull(groupName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"%s %s -> %s (%d):\n",
		GetEmojiByStatus(rule.State),
		groupName,
		rule.Name,
		len(rule.Alerts),
	))

	for _, alert := range rule.Alerts {
		for key, label := range alert.Labels {
			if key == "alertname" {
				continue
			}
			sb.WriteString(fmt.Sprintf("%s = %s\n", key, label))
		}
		if alert.Value != "" {
			sb.WriteString(fmt.Sprintf("value = %.1f\n", StrToFloat64(alert.Value)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (silence *Silence) Serialize() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:         <code>%s</code>\n", silence.ID))
	sb.WriteString(fmt.Sprintf("Starts at:  <pre>%s</pre>\n", silence.StartsAt.String()))
	sb.WriteString(fmt.Sprintf("Ends at:    <pre>%s</pre>\n", silence.EndsAt.String()))
	sb.WriteString(fmt.Sprintf("Created by: <pre>%s</pre>\n", silence.CreatedBy))
	sb.WriteString(fmt.Sprintf("Comment:    <pre>%s</pre>\n", silence.Comment))
	sb.WriteString(fmt.Sprintf(
		"Status:     <pre>%s %s</pre>\n",
		GetEmojiBySilenceStatus(silence.Status.State),
		silence.Status.State,
	))
	sb.WriteString("Matchers: ")

	for _, matcher := range silence.Matchers {
		sb.WriteString("<pre>" + matcher.Serialize() + "</pre> ")
	}

	return sb.String()
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
