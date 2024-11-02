package constants

const (
	SilenceMatcherRegexEqual    string = "=~"
	SilenceMatcherRegexNotEqual string = "!~"
	SilenceMatcherEqual         string = "="
	SilenceMatcherNotEqual      string = "!="

	SilencesInOneMessage             = 5
	AlertsInOneMessage               = 3
	GrafanaUnsilencePrefix           = "grafana_unsilence_"
	AlertmanagerUnsilencePrefix      = "alertmanager_unsilence_"
	GrafanaSilencePrefix             = "grafana_silence_"
	AlertmanagerSilencePrefix        = "alertmanager_silence_"
	GrafanaPrepareSilencePrefix      = "grafana_prepare_silence_"
	AlertmanagerPrepareSilencePrefix = "alertmanager_prepare_silence_"
)
