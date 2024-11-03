package utils

import (
	"main/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseRenderOptions(t *testing.T) {
	t.Parallel()

	result1, valid1 := ParseRenderOptions("/render PanelName")
	require.True(t, valid1)
	require.Equal(t, types.RenderOptions{
		Query:  "PanelName",
		Params: map[string]string{},
	}, result1)

	result2, valid2 := ParseRenderOptions("/render key1=value1 key2=value2 PanelName")
	require.True(t, valid2)
	require.Equal(t, types.RenderOptions{
		Query: "PanelName",
		Params: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}, result2)

	_, valid3 := ParseRenderOptions("/render key1=value1 key2=value2")
	require.False(t, valid3)

	_, valid4 := ParseRenderOptions("/render")
	require.False(t, valid4)
}

func TestSerializeQueryString(t *testing.T) {
	t.Parallel()

	qs := map[string]string{"key1": "value1", "key2": "value2"}
	serialized := SerializeQueryString(qs)
	require.Equal(t, "key1=value1&key2=value2", serialized)
}

func TestGetEmojiByStatus(t *testing.T) {
	t.Parallel()

	require.Equal(t, "🟢", GetEmojiByStatus("inactive"))
	require.Equal(t, "🟢", GetEmojiByStatus("ok"))
	require.Equal(t, "🟢", GetEmojiByStatus("normal"))
	require.Equal(t, "🟡", GetEmojiByStatus("pending"))
	require.Equal(t, "🔴", GetEmojiByStatus("firing"))
	require.Equal(t, "🔴", GetEmojiByStatus("alerting"))
	require.Equal(t, "[unknown]", GetEmojiByStatus("unknown"))
}

func TestGetEmojiBySilenceStatus(t *testing.T) {
	t.Parallel()

	require.Equal(t, "🟢", GetEmojiBySilenceStatus("active"))
	require.Equal(t, "⚪", GetEmojiBySilenceStatus("expired"))
	require.Equal(t, "[unknown]", GetEmojiBySilenceStatus("unknown"))
}

func TestParseSilenceFromCommand(t *testing.T) {
	t.Parallel()

	_, valid1 := ParseSilenceFromCommand("/silence", "sender")
	require.Equal(t, "Usage: /silence <duration> <params>", valid1)

	_, valid2 := ParseSilenceFromCommand("/silence invalid alertname", "sender")
	require.Equal(t, "Invalid duration provided!", valid2)

	silence3, valid3 := ParseSilenceFromCommand("/silence 2h comment=Comment key1=value1 key2!=value2 key3=~value3 key4!~value4 alertname", "sender")
	require.Empty(t, valid3)
	require.Equal(t, "Comment", silence3.Comment)
	require.Equal(t, "sender", silence3.CreatedBy)
	require.Equal(t, time.Now().Second(), silence3.StartsAt.Second())
	require.Equal(t, time.Now().Add(2*time.Hour).Second(), silence3.EndsAt.Second())
	require.Equal(t, types.SilenceMatchers{
		{IsEqual: true, IsRegex: false, Name: "key1", Value: "value1"},
		{IsEqual: false, IsRegex: false, Name: "key2", Value: "value2"},
		{IsEqual: true, IsRegex: true, Name: "key3", Value: "value3"},
		{IsEqual: false, IsRegex: true, Name: "key4", Value: "value4"},
		{IsEqual: true, IsRegex: false, Name: "alertname", Value: "alertname"},
	}, silence3.Matchers)

	_, valid4 := ParseSilenceFromCommand("/silence 48h comment=comment", "sender")
	require.Equal(t, "Usage: /silence <duration> <params>", valid4)
}

func TestParseSilenceWithDuration(t *testing.T) {
	t.Parallel()

	matchers := []types.QueryMatcher{{Key: "key", Operator: "unknown", Value: "value"}}
	silence, valid1 := ParseSilenceWithDuration("/silence", matchers, "sender", 2*time.Hour)
	require.Nil(t, silence)
	require.Equal(t, "Got unexpected operator: unknown", valid1)
}

func TestStrToFloat64Fail(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	StrToFloat64("invalid")
}

func TestStrToFloat64Ok(t *testing.T) {
	t.Parallel()

	value := StrToFloat64("0.123")
	require.InDelta(t, 0.123, value, 0.001)
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	duration := 26*time.Hour + 3*time.Minute
	value := FormatDuration(duration)
	require.Equal(t, "1 day 2 hours 3 minutes", value)
}

func TestFormatDate(t *testing.T) {
	t.Parallel()

	timezone, err := time.LoadLocation("Etc/GMT")
	require.NoError(t, err)

	date := time.Unix(0, 0)
	parsed := FormatDate(timezone)(date)
	require.Equal(t, "Thu, 01 Jan 1970 00:00:00 GMT", parsed)
}
