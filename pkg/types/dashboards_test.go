package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindDashboardByName(t *testing.T) {
	t.Parallel()

	dashboards := GrafanaDashboardsInfo{
		{Title: "dashboard"},
	}

	dashboard1, found1 := dashboards.FindDashboardByName("dashboard")
	require.NotNil(t, dashboard1)
	require.True(t, found1)

	dashboard2, found2 := dashboards.FindDashboardByName("unknown")
	require.Nil(t, dashboard2)
	require.False(t, found2)
}
