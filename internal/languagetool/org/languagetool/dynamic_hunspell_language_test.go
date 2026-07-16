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
}
