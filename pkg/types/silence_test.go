package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindSilence(t *testing.T) {
	t.Parallel()

	silences := Silences{
		{
			ID: "silence",
			Matchers: SilenceMatchers{
				{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
			},
		},
	}

	silence1, found1 := silences.FindByNameOrMatchers("silence")
	require.NotNil(t, silence1)
	require.True(t, found1)

	silence2, found2 := silences.FindByNameOrMatchers("key=value")
	require.NotNil(t, silence2)
	require.True(t, found2)

	silence3, found3 := silences.FindByNameOrMatchers("key1=value1 key2=value2")
	require.Nil(t, silence3)
	require.False(t, found3)
}

func TestSilenceGetFilterQueryString(t *testing.T) {
	t.Parallel()

	silence := Silence{
		ID: "silence",
		Matchers: SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	}

	qs := silence.GetFilterQueryString()
	require.Equal(t, "filter=key%3D%22value%22", qs)
}

func TestSilenceMatcherSerialize(t *testing.T) {
	t.Parallel()

	matcher := SilenceMatcher{IsEqual: true, IsRegex: false, Name: "key", Value: "value"}
	require.Equal(t, "key = value", matcher.Serialize())
	require.Equal(t, "key=\"value\"", matcher.SerializeQueryString())
}

func TestSilenceGetSymbol(t *testing.T) {
	t.Parallel()

	matcher1 := SilenceMatcher{IsEqual: true, IsRegex: false}
	require.Equal(t, "=", matcher1.GetSymbol())

	matcher2 := SilenceMatcher{IsEqual: true, IsRegex: true}
	require.Equal(t, "=~", matcher2.GetSymbol())

	matcher3 := SilenceMatcher{IsEqual: false, IsRegex: true}
	require.Equal(t, "!~", matcher3.GetSymbol())

	matcher4 := SilenceMatcher{IsEqual: false, IsRegex: false}
	require.Equal(t, "!=", matcher4.GetSymbol())
}

func TestSilenceMatchersEquals(t *testing.T) {
	t.Parallel()

	matchers := SilenceMatchers{{IsEqual: true, IsRegex: false, Name: "key", Value: "value"}}

	require.True(t, matchers.Equals(SilenceMatchers{
		{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
	}))
	require.False(t, matchers.Equals(SilenceMatchers{
		{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		{IsEqual: true, IsRegex: false, Name: "key2", Value: "value2"},
	}))
	require.False(t, matchers.Equals(SilenceMatchers{
		{IsEqual: true, IsRegex: false, Name: "key2", Value: "value2"},
	}))
}
