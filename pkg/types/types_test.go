package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryMatcherSerialize(t *testing.T) {
	t.Parallel()

	matcher := QueryMatcher{Key: "key", Operator: "=", Value: "value"}
	require.Equal(t, "key = value", matcher.Serialize())
}

func TestQueryMatcherFromKeyValueString(t *testing.T) {
	t.Parallel()

	require.Equal(t, QueryMatchers{
		{Key: "key", Operator: "=", Value: "value"},
	}, QueryMatcherFromKeyValueString("key=value"))

	require.Equal(t, QueryMatchers{
		{Key: "key", Operator: "=", Value: "value"},
	}, QueryMatcherFromKeyValueString("key=\"value\""))

	require.Equal(t, QueryMatchers{
		{Key: "key", Operator: "=", Value: "value"},
		{Key: "alertname", Operator: "=", Value: "alertname"},
	}, QueryMatcherFromKeyValueString("key=value alertname"))
}

func TestQueryMatcherFromKeyValueMap(t *testing.T) {
	t.Parallel()

	require.Equal(t, QueryMatchers{
		{Key: "key", Operator: "=", Value: "value"},
	}, QueryMatcherFromKeyValueMap(map[string]string{
		"key": "value",
	}))
}

func TestMatcherFromQueryMatcher(t *testing.T) {
	t.Parallel()

	require.Equal(t, &SilenceMatcher{
		IsEqual: true,
		IsRegex: false,
		Name:    "key",
		Value:   "value",
	}, MatcherFromQueryMatcher(QueryMatcher{Key: "key", Operator: "=", Value: "value"}))

	require.Equal(t, &SilenceMatcher{
		IsEqual: true,
		IsRegex: true,
		Name:    "key",
		Value:   "value",
	}, MatcherFromQueryMatcher(QueryMatcher{Key: "key", Operator: "=~", Value: "value"}))

	require.Equal(t, &SilenceMatcher{
		IsEqual: false,
		IsRegex: true,
		Name:    "key",
		Value:   "value",
	}, MatcherFromQueryMatcher(QueryMatcher{Key: "key", Operator: "!~", Value: "value"}))

	require.Equal(t, &SilenceMatcher{
		IsEqual: false,
		IsRegex: false,
		Name:    "key",
		Value:   "value",
	}, MatcherFromQueryMatcher(QueryMatcher{Key: "key", Operator: "!=", Value: "value"}))
}
