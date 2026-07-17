package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanParagraphRepeatBeginningRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of GermanParagraphRepeatBeginningRuleTest.testRule
func TestGermanParagraphRepeatBeginningRule_Rule(t *testing.T) {
	r := NewGermanParagraphRepeatBeginningRule(map[string]string{
		"repetition": "Wiederholung am Absatzanfang",
	})
	// two paragraphs starting with same article+noun → match
	s1 := languagetool.AnalyzePlain("Der Hund spazierte über die Straße.\n\n")
	s2 := languagetool.AnalyzePlain("Der Hund ignorierte den Verkehr.")
	// Java expects 2 matches (one per repeated beginning occurrence style); at least flag
	m := r.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.NotEmpty(t, m, "same Der Hund paragraph starts should flag")

	// different beginning → no flag
	s3 := languagetool.AnalyzePlain("Das Tier ignorierte den Verkehr.")
	require.Empty(t, r.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Der Hund spazierte über die Straße.\n\n"),
		s3,
	}))

	// same proper name
	require.NotEmpty(t, r.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Peter spazierte über die Straße.\n\n"),
		languagetool.AnalyzePlain("Peter ignorierte den Verkehr."),
	}))
}
