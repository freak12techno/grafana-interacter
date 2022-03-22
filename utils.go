package main

import (
	"regexp"
	"strings"
)

func NormalizeString(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(input, ""))
}

func FindDashboardByName(dashboards []GrafanaDashboardInfo, name string) (*GrafanaDashboardInfo, bool) {
	normalizedName := NormalizeString(name)

	for _, dashboard := range dashboards {
		if strings.Contains(NormalizeString(dashboard.Title), normalizedName) {
			return &dashboard, true
		}
	}

	return nil, false
}

func FindPanelByName(panels []PanelStruct, name string) (*PanelStruct, bool) {
	normalizedName := NormalizeString(name)

	for _, panel := range panels {
		panelNameWithDashboardName := NormalizeString(panel.DashboardName + panel.Name)

		if strings.Contains(panelNameWithDashboardName, normalizedName) {
			return &panel, true
		}
	}

	return nil, false
}

func ParseRenderOptions(query string) (RenderOptions, bool) {
	args := strings.Split(query, " ")
	if len(args) <= 1 {
		return RenderOptions{}, false // should have at least 1 argument
	}

	params := map[string]string{}

	_, args = args[0], args[1:] // removing first argument as it's always /render
	for len(args) > 0 {
		if !strings.Contains(args[0], "=") {
			break
		}

		paramSplit := strings.SplitN(args[0], "=", 2)
		params[paramSplit[0]] = paramSplit[1]

		_, args = args[0], args[1:]
	}

	return RenderOptions{
		Query:  strings.Join(args, " "),
		Params: params,
	}, len(args) > 0
}

func SerializeQueryString(qs map[string]string) string {
	tmp := make([]string, len(qs))
	counter := 0

	for key, value := range qs {
		tmp[counter] = key + "=" + value
		counter++
	}

	return strings.Join(tmp, "&")
}

func MergeMaps(first, second map[string]string) map[string]string {
	for key, value := range second {
		first[key] = value
	}

	return first
}
