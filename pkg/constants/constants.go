package constants

const (
	SilenceMatcherRegexEqual    string = "=~"
	SilenceMatcherRegexNotEqual string = "!~"
	SilenceMatcherEqual         string = "="
	SilenceMatcherNotEqual      string = "!="

	SilencesInOneMessage        = 5
	GrafanaUnsilencePrefix      = "unsilence"
	AlertmanagerUnsilencePrefix = "alertmanager_unsilence"
)
