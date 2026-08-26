package suggestions_ordering

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func resetMLOrderingProperty(t *testing.T) {
	t.Helper()
	clearSystemProperty(propName)
	t.Cleanup(func() {
		clearSystemProperty(propName)
	})
}

func TestSuggestionsOrdererConfig_DefaultDisabled(t *testing.T) {
	resetMLOrderingProperty(t)
	require.False(t, IsMLSuggestionsOrderingEnabled())
}

func TestSuggestionsOrdererConfig_EnableDisable(t *testing.T) {
	resetMLOrderingProperty(t)
	require.False(t, IsMLSuggestionsOrderingEnabled())

	SetMLSuggestionsOrderingEnabled(true)
	require.True(t, IsMLSuggestionsOrderingEnabled())

	SetMLSuggestionsOrderingEnabled(false)
	require.False(t, IsMLSuggestionsOrderingEnabled())
}

func TestSuggestionsOrdererConfig_ParseBooleanQuirks(t *testing.T) {
	resetMLOrderingProperty(t)

	// missing key → default "false" → parseBoolean → false
	require.False(t, IsMLSuggestionsOrderingEnabled())

	// Boolean.parseBoolean: only "true" (case-insensitive) is true
	setSystemProperty(propName, "true")
	require.True(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "TRUE")
	require.True(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "True")
	require.True(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "tRuE")
	require.True(t, IsMLSuggestionsOrderingEnabled())

	// "false", garbage, empty → false (unlike strconv.ParseBool which accepts "1","t",…)
	setSystemProperty(propName, "false")
	require.False(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "FALSE")
	require.False(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "yes")
	require.False(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "1")
	require.False(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "t")
	require.False(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "")
	require.False(t, IsMLSuggestionsOrderingEnabled())
	setSystemProperty(propName, "garbage")
	require.False(t, IsMLSuggestionsOrderingEnabled())
}

func TestSuggestionsOrdererConfig_SetStoresStringValueOf(t *testing.T) {
	resetMLOrderingProperty(t)

	SetMLSuggestionsOrderingEnabled(true)
	require.Equal(t, "true", getSystemProperty(propName, ""))
	SetMLSuggestionsOrderingEnabled(false)
	require.Equal(t, "false", getSystemProperty(propName, ""))
}

func TestSuggestionsOrdererConfig_NgramsPathRoundTrip(t *testing.T) {
	prev := GetNgramsPath()
	t.Cleanup(func() {
		SetNgramsPath(prev)
	})

	SetNgramsPath("/tmp/ngrams")
	require.Equal(t, "/tmp/ngrams", GetNgramsPath())
	SetNgramsPath("")
	require.Equal(t, "", GetNgramsPath())
	SetNgramsPath("/other/path")
	require.Equal(t, "/other/path", GetNgramsPath())
}

func TestParseBoolean(t *testing.T) {
	require.True(t, parseBoolean("true"))
	require.True(t, parseBoolean("TRUE"))
	require.True(t, parseBoolean("True"))
	require.False(t, parseBoolean("false"))
	require.False(t, parseBoolean(""))
	require.False(t, parseBoolean("1"))
	require.False(t, parseBoolean("t"))
	require.False(t, parseBoolean("yes"))
}
