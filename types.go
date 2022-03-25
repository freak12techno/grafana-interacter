package main

import "fmt"

type ConfigStruct struct {
	LogLevel   string         `yaml:"log_level" default:"info"`
	JSONOutput bool           `yaml:"json" default:"false"`
	Telegram   TelegramConfig `yaml:"telegram" default:""`
	Grafana    GrafanaConfig  `yaml:"grafana"`
}

type TelegramConfig struct {
	Token  string  `yaml:"token"`
	Admins []int64 `yaml:"admins"`
}

type GrafanaConfig struct {
	URL      string `yaml:"url" default:"http://localhost:3000"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

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
}

type RenderOptions struct {
	Query  string
	Params map[string]string
}

func (rule *GrafanaAlertRule) Serialize(groupName string) string {
	return fmt.Sprintf("- %s %s -> %s\n", GetEmojiByStatus(rule.State), groupName, rule.Name)
}

func (alert *GrafanaAlert) Serialize() string {
	return fmt.Sprintf("- %s <pre>%s</pre>", GetEmojiByStatus(alert.State), SerializeAlertLabels(alert.Labels))
}
