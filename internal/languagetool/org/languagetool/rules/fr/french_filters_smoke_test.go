package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFrenchNumberInWordFilter_Suggestions(t *testing.T) {
	ClearFrenchFilterSpeller()
	f := NewFrenchNumberInWordFilter()
	// fail-closed without speller (Java isMisspelled gate)
	require.Empty(t, f.Suggestions("m0t"))
	f.inner.IsMisspelled = func(w string) bool { return w != "mot" && w != "mt" }
	f.inner.GetSuggestions = nil
	sugg := f.inner.Suggestions("m0t")
	require.Contains(t, sugg, "mot")
}

func TestFrenchRepeatedWordsRule_Construct(t *testing.T) {
	r := NewFrenchRepeatedWordsRule(nil)
	require.NotNil(t, r)
	// MatchList may need synonym data — ensure call is safe
	_ = r.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Le grand grand livre.")})
}

func TestFrenchCommaWhitespace_Construct(t *testing.T) {
	r := NewCommaWhitespaceRule(nil)
	require.NotNil(t, r)
	ms := r.Match(languagetool.AnalyzePlain("Bonjour , monde"))
	// may flag space before comma
	_ = ms
}
