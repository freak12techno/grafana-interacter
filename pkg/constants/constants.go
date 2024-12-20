package constants

const (
	SilenceMatcherRegexEqual    string = "=~"
	SilenceMatcherRegexNotEqual string = "!~"
	SilenceMatcherEqual         string = "="
	SilenceMatcherNotEqual      string = "!="

	SilencesInOneMessage   = 5
	AlertsInOneMessage     = 3
	DashboardsInOneMessage = 5
	PanelsInOneMessage     = 5

	GrafanaPaginatedFiringAlertsList    = "grafana_paginated_firing_alerts_list_"
	PrometheusPaginatedFiringAlertsList = "prometheus_paginated_firing_alerts_list_"
	GrafanaPaginatedSilencesList        = "grafana_paginated_silences_list_"
	AlertmanagerPaginatedSilencesList   = "alertmanager_paginated_silences_list_"
	GrafanaUnsilencePrefix              = "grafana_unsilence_"
	AlertmanagerUnsilencePrefix         = "alertmanager_unsilence_"
	GrafanaSilencePrefix                = "grafana_silence_"
	AlertmanagerSilencePrefix           = "alertmanager_silence_"
	GrafanaPrepareSilencePrefix         = "grafana_prepare_silence_"
	AlertmanagerPrepareSilencePrefix    = "alertmanager_prepare_silence_"
	GrafanaListSilencesCommand          = "grafana_silences"
	AlertmanagerListSilencesCommand     = "alertmanager_silences"
	GrafanaSilenceCommand               = "grafana_silence"
	AlertmanagerSilenceCommand          = "alertmanager_silence"
	GrafanaUnsilenceCommand             = "grafana_unsilence"
	AlertmanagerUnsilenceCommand        = "alertmanager_unsilence"

	GrafanaRenderChooseDashboardPrefix = "render_choose_dashboard_"
	GrafanaRenderChoosePanelPrefix     = "render_choose_panel_"
	GrafanaRenderRenderPanelPrefix     = "render_render_panel"
	ClearKeyboardPrefix                = "clear_keyboard_"
)
