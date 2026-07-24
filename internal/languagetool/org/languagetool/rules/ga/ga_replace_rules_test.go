package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishHomophoneRule(t *testing.T) {
	rule := NewEnglishHomophoneRule(nil)
	// Java example: sushi → súisí
	matches := rule.Match(languagetool.AnalyzePlain("An bhialann sushi sin ba chúis leis."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "súisí", matches[0].GetSuggestedReplacements()[0])
}

func TestIrishFGBEqReplaceRule(t *testing.T) {
	rule := NewIrishFGBEqReplaceRule(nil)
	// Java example: urlamh → ullamh
	matches := rule.Match(languagetool.AnalyzePlain("An bhfuil tú urlamh?"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "ullamh", matches[0].GetSuggestedReplacements()[0])
}

func TestPrestandardReplaceRule(t *testing.T) {
	rule := NewPrestandardReplaceRule(nil)
	// Java example: baoghal → baol
	matches := rule.Match(languagetool.AnalyzePlain("Ní baoghal daoibh."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "baol", matches[0].GetSuggestedReplacements()[0])
}

func TestPeopleRule(t *testing.T) {
	rule := NewPeopleRule(nil)
	// Java example: Damocles → Dámaicléas
	matches := rule.Match(languagetool.AnalyzePlain("claíomh Damocles ar crochadh"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Dámaicléas", matches[0].GetSuggestedReplacements()[0])
}

func TestIrishSpecificCaseRule(t *testing.T) {
	rule := NewIrishSpecificCaseRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Béal Feirste"))))
	// Java example: mbéal Feirste → mBéal Feirste
	matches := rule.Match(languagetool.AnalyzePlain("i mbéal Feirste é."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "mBéal Feirste", matches[0].GetSuggestedReplacements()[0])
}
