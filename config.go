package main

type ConfigStruct struct {
	LogLevel      string        `yaml:"log_level" default:"info"`
	JSONOutput    bool          `yaml:"json" default:"false"`
	TelegramToken string        `yaml:"telegram_token" default:""`
	Auth          AuthStruct    `yaml:"auth"`
	GrafanaURL    string        `yaml:"grafana_url" default:"http://localhost:3000"`
	Panels        []PanelStruct `yaml:"panels"`
}

type AuthStruct struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type PanelStruct struct {
	Name        string `yaml:"name"`
	DashboardID string `yaml:"dashboard_id"`
	PanelID     string `yaml:"panel_id"`
}
