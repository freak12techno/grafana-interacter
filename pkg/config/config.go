package config

type Config struct {
	Log          LogConfig          `yaml:"log"`
	Telegram     TelegramConfig     `default:""          yaml:"telegram"`
	Grafana      GrafanaConfig      `yaml:"grafana"`
	Alertmanager AlertmanagerConfig `yaml:"alertmanager"`
}

type LogConfig struct {
	LogLevel   string `default:"info"  yaml:"level"`
	JSONOutput bool   `default:"false" yaml:"json"`
}

type TelegramConfig struct {
	Token  string  `yaml:"token"`
	Admins []int64 `yaml:"admins"`
}

type GrafanaConfig struct {
	URL           string            `default:"http://localhost:3000" yaml:"url"`
	User          string            `yaml:"user"`
	Password      string            `yaml:"password"`
	RenderOptions map[string]string `yaml:"render_options" default:"{\"orgId\":\"1\",\"from\":\"now\",\"to\":\"now-30m\"}"`
}

type AlertmanagerConfig struct {
	URL      string `default:"http://localhost:9093" yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Timezone string `default:"Europe/Moscow"         yaml:"timezone"`
}
