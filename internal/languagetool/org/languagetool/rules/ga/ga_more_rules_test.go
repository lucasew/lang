package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestLogainmRule(t *testing.T) {
	rule := NewLogainmRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("I live in Dublin."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Baile Átha Cliath", matches[0].GetSuggestedReplacements()[0])
}

func TestSpacesRule(t *testing.T) {
	rule := NewSpacesRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("rud abhfuil ann"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "a bhfuil", matches[0].GetSuggestedReplacements()[0])
}

func TestDativePluralStandardReplaceRule(t *testing.T) {
	rule := NewDativePluralStandardReplaceRule(nil)
	// Java example: mnáibh → mná
	matches := rule.Match(languagetool.AnalyzePlain("do na mnáibh a gcás"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "mná", matches[0].GetSuggestedReplacements()[0])
}
