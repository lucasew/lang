package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanParagraphRepeatBeginningRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of GermanParagraphRepeatBeginningRuleTest.testRule
func TestGermanParagraphRepeatBeginningRule_Rule(t *testing.T) {
	r := NewGermanParagraphRepeatBeginningRule(map[string]string{
		"repetition": "Wiederholung am Absatzanfang",
	})
	// Java isArticle is ART POS only — inject ART for Der/Das (no surface invent).
	// two paragraphs starting with same article+noun → match (2 spans)
	s1 := withART("Der Hund spazierte über die Straße.\n\n", "Der")
	s2 := withART("Der Hund ignorierte den Verkehr.", "Der")
	m := r.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.NotEmpty(t, m, "same Der Hund paragraph starts should flag")

	// different beginning → no flag
	s3 := withART("Das Tier ignorierte den Verkehr.", "Das")
	require.Empty(t, r.MatchList([]*languagetool.AnalyzedSentence{
		withART("Der Hund spazierte über die Straße.\n\n", "Der"),
		s3,
	}))

	// same proper name (not article)
	require.NotEmpty(t, r.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Peter spazierte über die Straße.\n\n"),
		languagetool.AnalyzePlain("Peter ignorierte den Verkehr."),
	}))
}

func TestGermanParagraphRepeatBeginningRule_IsArticlePOSOnly(t *testing.T) {
	r := NewGermanParagraphRepeatBeginningRule(nil)
	// Untagged "Der" is not an article (Java ART only).
	tok := languagetool.AnalyzePlain("Der Hund").GetTokensWithoutWhitespace()[1]
	require.False(t, r.IsArticle(tok))
	// Inject ART
	pos := "ART:DEF:Nom:Sg:Masc"
	tok.AddReading(languagetool.NewAnalyzedToken("Der", &pos, nil), "test")
	require.True(t, r.IsArticle(tok))
}

func withART(text, article string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		if strings.EqualFold(tok, article) {
			return []languagetool.TokenTag{{POS: "ART:DEF", Lemma: strings.ToLower(tok)}}
		}
		return nil
	})
}
