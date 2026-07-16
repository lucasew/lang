package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleRepeatedSentenceBeginning(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// Java example: three sentences starting with articles
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund lief langsam über die Straße."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 3, len(matches))

	// mixed starts — no streak of 3
	sents2 := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Langsam lief der Hund."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	require.Equal(t, 0, len(rule.MatchList(sents2)))
}
