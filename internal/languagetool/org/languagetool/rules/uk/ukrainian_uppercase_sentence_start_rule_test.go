package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkrainianUppercaseSentenceStartRule(t *testing.T) {
	rule := NewUkrainianUppercaseSentenceStartRule(nil)
	// list item exception
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("а) перший пункт"),
	})))
	// lowercase start still flagged when multi-token sentence
	// (single-word short sentences are ignored by the core rule)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("це звичайне речення з кількома словами."),
	})
	require.Equal(t, 1, len(matches))
}
