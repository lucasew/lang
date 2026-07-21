package de

// Twin of GermanUnpairedQuotesRuleTest — Java uses de-DE JLanguageTool (GermanWordTokenizer).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
	"github.com/stretchr/testify/require"
)

func deAnalyzeUQ(text string) *languagetool.AnalyzedSentence {
	// Java: lt.getAnalyzedSentence → language word tokenizer (GermanWordTokenizer adds ‚).
	return languagetool.AnalyzeWithTokenizer(text, detok.NewGermanWordTokenizer())
}

func TestGermanUnpairedQuotesRule_GermanRule(t *testing.T) {
	rule := NewGermanUnpairedQuotesRule(nil)
	assertMatches := func(input string, expected int) {
		t.Helper()
		n := len(rule.MatchList([]*languagetool.AnalyzedSentence{deAnalyzeUQ(input)}))
		require.Equal(t, expected, n, "input=%q", input)
	}

	// correct sentences (Java):
	assertMatches("»Das sind die Sätze, die sie testen sollen«.", 0)
	assertMatches("«Das sind die ‹Sätze›, die sie testen sollen».", 0)
	assertMatches("»Das sind die ›Sätze‹, die sie testen sollen«.", 0)
	assertMatches("»Das sind die Sätze ›noch mehr Anführungszeichen‹ ›schon wieder!‹, die sie testen sollen«.", 0)
	assertMatches("»Das sind die Sätze ›noch mehr Anführungszeichen ›hier ein Fehler!‹‹, die sie testen sollen«.", 2)
	assertMatches("„Das sind die Sätze ‚noch mehr Anführungszeichen‘ ‚schon wieder!‘, die sie testen sollen“.", 0)
	assertMatches("„Das sind die Sätze ‚noch mehr Anführungszeichen ‚hier ein Fehler!‘‘, die sie testen sollen“.", 2)
	assertMatches("„Das sind die Sätze, die sie testen sollen.“ „Hier steht ein zweiter Satz.“", 0)
	assertMatches("Drücken Sie auf den \"Jetzt Starten\"-Knopf.", 0)
	assertMatches("Welches ist dein Lieblings-\"Star Wars\"-Charakter?", 0)
	assertMatches("‚So 'n Blödsinn!‘", 0)
	assertMatches("‚’n Blödsinn!‘", 0)
	assertMatches("'So 'n Blödsinn!'", 0)
	assertMatches("''n Blödsinn!'", 0)
	assertMatches("‚Das ist Hans’.‘", 0)
	assertMatches("'Das ist Hans'.'", 0)
	assertMatches("Das Fahrrad hat 26\" Räder.", 0)
	assertMatches("\"Das Fahrrad hat 26\" Räder.\"", 0)
	assertMatches("und steigern » Datenbankperformance steigern » Tipps zur Performance-Verbesserung", 0)

	// incorrect:
	assertMatches("\"Das Fahrrad hat 26\" Räder.\" \"Und hier fehlt das abschließende doppelte Anführungszeichen.", 1)
	assertMatches("Die „Sätze zum Testen.", 1)
	assertMatches("Die «Sätze zum Testen.", 1)
	assertMatches("Die »Sätze zum Testen.", 1)

	require.Equal(t, "DE_UNPAIRED_QUOTES", rule.GetID())
	require.Contains(t, rule.GetURL(), "klammern")
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{deAnalyzeUQ("Die „Sätze zum Testen.")})
	require.NotEmpty(t, ms)
	require.Equal(t, rule, ms[0].GetRule())

	// Java: soft-hyphen sentences that used to break position mapping — smoke (no panic).
	// Full JLT check path maps soft hyphens via AnnotatedText; here MatchList must not panic.
	for _, s := range []string{
		"Im Kran\u00ADken\u00ADhaus. Auch)",
		"Ein Kran\u00ADken\u00ADhaus. Auch)",
		"Das Kran\u00ADken\u00ADhaus. Auch)",
		"Kran\u00ADken\u00ADhaus. Auch)",
		"Kran\u00ADken\u00ADhaus. (Auch",
	} {
		require.NotPanics(t, func() {
			_ = rule.MatchList([]*languagetool.AnalyzedSentence{deAnalyzeUQ(s)})
		}, s)
	}
}
