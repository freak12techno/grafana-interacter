package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetAlertSerializeLabels(t *testing.T) {
	t.Parallel()
	alert := GrafanaAlert{
		Labels: map[string]string{
			"key2": "value2",
			"key1": "value1",
			"key3": "value3",
		},
	}

	serialized := alert.SerializeLabels()
	require.Equal(t, "key1=value1 key2=value2 key3=value3", serialized)
}

func TestGetAlertCallbackHash(t *testing.T) {
	t.Parallel()
	alert := GrafanaAlert{
		Labels: map[string]string{
			"key2": "value2",
			"key1": "value1",
			"key3": "value3",
		},
	}

	hash := alert.GetCallbackHash()
	require.Equal(t, "4d92bd65991b9a6c6fb16973356352b1", hash)
}

func TestGetAlertGetActiveSinc(t *testing.T) {
	t.Parallel()

	alert := GrafanaAlert{ActiveAt: time.Now().Add(-1 * time.Hour)}
	since := alert.ActiveSince()
	require.Equal(t, int(1*time.Hour.Seconds()), int(since.Seconds()))
}

func TestFindAlertRuleByName(t *testing.T) {
	t.Parallel()

	groups := GrafanaAlertGroups{
		{
			Name: "group",
			Rules: []GrafanaAlertRule{
				{
					Name: "rule",
				},
			},
		},
	}

	alert1, found1 := groups.FindAlertRuleByName("rule")
	require.NotNil(t, alert1)
	require.True(t, found1)

	alert2, found2 := groups.FindAlertRuleByName("unknown")
	require.Nil(t, alert2)
	require.False(t, found2)
}

func TestFindLabelsByHash(t *testing.T) {
	t.Parallel()

	groups := GrafanaAlertGroups{
		{
			Name: "group",
			Rules: []GrafanaAlertRule{
				{
					Name: "rule",
					Alerts: []GrafanaAlert{
						{
							Labels: map[string]string{
								"key2": "value2",
								"key1": "value1",
								"key3": "value3",
							},
						},
					},
				},
			},
		},
	}

	labels1, found1 := groups.FindLabelsByHash("4d92bd65991b9a6c6fb16973356352b1")
	require.NotNil(t, labels1)
	require.True(t, found1)

	labels2, found2 := groups.FindLabelsByHash("unknown")
	require.Nil(t, labels2)
	require.False(t, found2)
}

func TestFilterFiringOrPendingGrous(t *testing.T) {
	t.Parallel()

	groups := GrafanaAlertGroups{
		{
			Name: "group",
			Rules: []GrafanaAlertRule{
				{
					Name:  "rule",
					State: "not-firing",
				},
				{
					Name:  "rule",
					State: "firing",
					Alerts: []GrafanaAlert{
						{State: "not-firing"},
					},
				},
				{
					Name:  "rule",
					State: "firing",
					Alerts: []GrafanaAlert{
						{State: "not-firing"},
						{State: "firing"},
					},
				},
			},
		},
	}

	groups = groups.FilterFiringOrPendingAlertGroups(true)

	require.Len(t, groups, 1)

	group := groups[0]
	require.Len(t, group.Rules, 1)

	rule := group.Rules[0]
	require.Len(t, rule.Alerts, 1)
}
