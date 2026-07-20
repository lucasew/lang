package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamicMorfologikLanguage(t *testing.T) {
	d := NewDynamicMorfologikLanguage("Custom", "en-US", "/tmp/dict.dict")
	require.Equal(t, "en", d.GetShortCode())
	require.Equal(t, "EN-US_SPELLER_RULE", d.SpellerRuleID())
	require.Equal(t, []string{"EN-US_SPELLER_RULE"}, d.RelevantSpellerRuleIDs())
	require.Equal(t, "/tmp/dict.dict", d.GetFileName())
	require.Nil(t, d.GetSpellingFileName())
}
