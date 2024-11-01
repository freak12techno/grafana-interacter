package constants

const (
	SilenceMatcherRegexEqual    string = "=~"
	SilenceMatcherRegexNotEqual string = "!~"
	SilenceMatcherEqual         string = "="
	SilenceMatcherNotEqual      string = "!="

	SilencesInOneMessage        = 5
	AlertsInOneMessage          = 3
	GrafanaUnsilencePrefix      = "unsilence_"
	AlertmanagerUnsilencePrefix = "alertmanager_unsilence_"
)
