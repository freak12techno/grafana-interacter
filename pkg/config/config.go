package config

import (
	"fmt"
	"time"
)

type Config struct {
	Timezone     string             `default:"Etc/GMT"   yaml:"timezone"`
	Log          LogConfig          `yaml:"log"`
	Telegram     TelegramConfig     `yaml:"telegram"`
	Grafana      GrafanaConfig      `yaml:"grafana"`
	Alertmanager AlertmanagerConfig `yaml:"alertmanager"`
	Prometheus   *PrometheusConfig  `yaml:"prometheus"`
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
	URL            string            `default:"http://localhost:3000"                                 yaml:"url"`
	User           string            `default:"admin"                                                 yaml:"user"`
	Password       string            `default:"admin"                                                 yaml:"password"`
	Token          string            `yaml:"token"`
	RenderOptions  map[string]string `default:"{\"orgId\":\"1\",\"from\":\"now\",\"to\":\"now-30m\"}" yaml:"render_options"`
	MutesDurations []string          `default:"[\"1h\",\"8h\",\"24h\",\"168h\",\"99999h\"]"           yaml:"mutes_durations"`
}

type PrometheusConfig struct {
	URL      string `default:"http://localhost:9090" yaml:"url"`
	User     string `default:"admin"                 yaml:"user"`
	Password string `default:"admin"                 yaml:"password"`
}

type AlertmanagerConfig struct {
	URL            string   `default:"http://localhost:9093"                       yaml:"url"`
	User           string   `yaml:"user"`
	Password       string   `yaml:"password"`
	MutesDurations []string `default:"[\"1h\",\"8h\",\"24h\",\"168h\",\"99999h\"]" yaml:"mutes_durations"`
}

func (c *Config) Validate() error {
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("error parsing timezone: %s", err)
	}

	return nil
}
