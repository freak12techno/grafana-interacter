package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindPanel(t *testing.T) {
	t.Parallel()

	panels := PanelsStruct{
		{DashboardName: "dashboard", Name: "panel"},
	}

	panel1, found1 := panels.FindByName("dashboard panel")
	require.NotNil(t, panel1)
	require.True(t, found1)

	panel2, found2 := panels.FindByName("panel")
	require.NotNil(t, panel2)
	require.True(t, found2)

	panel3, found3 := panels.FindByName("unknown")
	require.Nil(t, panel3)
	require.False(t, found3)
}
