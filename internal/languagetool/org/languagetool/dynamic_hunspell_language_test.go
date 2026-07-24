package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamicHunspellLanguage(t *testing.T) {
	d := NewDynamicHunspellLanguage("Foo", "xx-YY", "/path/to/dict.dic")
	require.Equal(t, "XX-YY_SPELLER_RULE", d.SpellerRuleID())
	require.Equal(t, "/path/to/dict", d.DictFilenameInResources())
	require.Equal(t, "xx", d.GetShortCode())
	require.Nil(t, d.GetSpellingFileName())
	require.Equal(t, []string{"XX-YY_SPELLER_RULE"}, d.RelevantSpellerRuleIDs())
	// Java regex .dic$ strips any-char+"dic" at end
	d2 := NewDynamicHunspellLanguage("F", "aa", "/x/foodic")
	require.Equal(t, "/x/fo", d2.DictFilenameInResources()) // regex .dic$ eats any+dic
}
