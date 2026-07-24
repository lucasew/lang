package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceDiacriticsIEC(t *testing.T) {
	rule := NewSimpleReplaceDiacriticsIEC(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Va dir adéu i va marxar."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "adeu", matches[0].GetSuggestedReplacements()[0])
}

func TestSimpleReplaceAdverbsMent(t *testing.T) {
	rule := NewSimpleReplaceAdverbsMent(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Ho farà ràpidament."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "de pressa", matches[0].GetSuggestedReplacements()[0])
}
