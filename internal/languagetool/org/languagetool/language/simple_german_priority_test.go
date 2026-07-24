package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleGermanPriorityForId(t *testing.T) {
	// Java SimpleGerman.getPriorityForId
	require.Equal(t, 10, SimpleGermanPriorityForId("TOO_LONG_SENTENCE"))
	require.Equal(t, -1, SimpleGermanPriorityForId("LANGES_WORT"))
	// super German
	require.Equal(t, 10, SimpleGermanPriorityForId("OLD_SPELLING_RULE"))
	require.Equal(t, 11, SimpleGermanPriorityForId("DE_PROHIBITED_PHRASE"))
	// code selection
	sg := GermanPriorityForIdForCode("de-DE-x-simple-language")
	require.Equal(t, 10, sg("TOO_LONG_SENTENCE"))
	de := GermanPriorityForIdForCode("de-DE")
	// base German: TOO_LONG_SENTENCE is Language base -101, not 10
	require.Equal(t, -101, de("TOO_LONG_SENTENCE"))
	require.True(t, isSimpleGermanCode("de-DE-x-simple-language"))
	require.False(t, isSimpleGermanCode("de-DE"))
}

func TestSimpleGerman_GetRuleFileNames(t *testing.T) {
	// Java SimpleGerman.getRuleFileNames: only private-use shortCode/grammar.xml (no super)
	require.Equal(t, []string{
		"/org/languagetool/rules/de-DE-x-simple-language/grammar.xml",
	}, SimpleGermanGetRuleFileNames())
	require.Equal(t, "de-DE-x-simple-language", SimpleGermanShortCode)
	require.True(t, SimpleGermanIsVariant())
	require.Equal(t, "Simple German", SimpleGermanGetName())
	ms := SimpleGermanGetMaintainers()
	require.Len(t, ms, 1)
	require.Equal(t, "Annika Nietzio", ms[0].Name)
}
