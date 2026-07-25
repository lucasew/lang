package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanUnpairedQuotesRuleTest.java
//
// Java setup:
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   rule = GenericUnpairedQuotesRuleTest.getQuotesRule(lt); // active GenericUnpairedQuotesRule
//   assertMatches: rule.match(Collections.singletonList(lt.getAnalyzedSentence(input))).length
// Soft-hyphen lt.check(...) calls only assert no throw / position mapping.
//
// Go: NewGermanUnpairedQuotesRule + de-DE JLanguageTool.GetAnalyzedSentence (real analysis).
// DE quotes do not need EN-style apostrophe POS invent — base is surface + isNonWord.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// TestGermanUnpairedQuotesRule_GermanRule ports GermanUnpairedQuotesRuleTest.testGermanRule.
func TestGermanUnpairedQuotesRule_GermanRule(t *testing.T) {
	// Java: getQuotesRule finds GermanUnpairedQuotesRule on de-DE; construct directly.
	rule := NewGermanUnpairedQuotesRule(nil)
	lt := languagetool.NewJLanguageTool("de-DE")

	assertMatches := func(input string, expected int) {
		t.Helper()
		// Java: rule.match(Collections.singletonList(lt.getAnalyzedSentence(input)))
		sent := lt.GetAnalyzedSentence(input)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", input)
		n := len(rule.MatchList([]*languagetool.AnalyzedSentence{sent}))
		require.Equal(t, expected, n, "input=%q", input)
	}

	// correct sentences (Java assertMatches … 0 / 2):
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

	// incorrect sentences (Java):
	assertMatches("\"Das Fahrrad hat 26\" Räder.\" \"Und hier fehlt das abschließende doppelte Anführungszeichen.", 1)
	assertMatches("Die „Sätze zum Testen.", 1)
	assertMatches("Die «Sätze zum Testen.", 1)
	assertMatches("Die »Sätze zum Testen.", 1)

	// Meta: id / URL / match.Rule attachment (Java setUrl on this DE rule)
	require.Equal(t, "DE_UNPAIRED_QUOTES", rule.GetID())
	require.Equal(t, "https://languagetool.org/insights/de/beitrag/klammern/", rule.GetURL())
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{lt.GetAnalyzedSentence("Die „Sätze zum Testen.")})
	require.NotEmpty(t, ms)
	require.Equal(t, rule, ms[0].GetRule())

	// Example pair from Java constructor
	require.Len(t, rule.GetIncorrectExamples(), 1)
	require.Len(t, rule.GetCorrectExamples(), 1)
	require.Equal(t, []string{"›"}, rule.GetIncorrectExamples()[0].GetCorrections())

	// Start/end symbols (Java DE_START_SYMBOLS / DE_END_SYMBOLS)
	require.Equal(t, []string{"„", "»", "«", "\"", "'", "‚", "›", "‹"}, rule.StartSymbols)
	require.Equal(t, []string{"“", "«", "»", "\"", "'", "‘", "‹", "›"}, rule.EndSymbols)

	// Java: soft-hyphen sentences that used to break position mapping — no throw.
	// Full JLT check path maps soft hyphens via AnnotatedText; MatchList / Check must not panic.
	for _, s := range []string{
		"Im Kran\u00ADken\u00ADhaus. Auch)",
		"Ein Kran\u00ADken\u00ADhaus. Auch)",
		"Das Kran\u00ADken\u00ADhaus. Auch)",
		"Kran\u00ADken\u00ADhaus. Auch)",
		"Kran\u00ADken\u00ADhaus. (Auch",
	} {
		require.NotPanics(t, func() {
			_ = rule.MatchList([]*languagetool.AnalyzedSentence{lt.GetAnalyzedSentence(s)})
		}, "MatchList soft-hyphen %q", s)
		require.NotPanics(t, func() {
			_ = lt.Check(s)
		}, "Check soft-hyphen %q", s)
	}
}
