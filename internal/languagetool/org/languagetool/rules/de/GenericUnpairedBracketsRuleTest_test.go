package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GenericUnpairedBracketsRuleTest.java
//
// Java setup:
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   rule = GenericUnpairedBracketsRuleTest.getBracketsRule(lt); // active GenericUnpairedBracketsRule
//   assertMatches: rule.match(Collections.singletonList(lt.getAnalyzedSentence(input))).length
// Soft-hyphen lt.check(...) calls only assert no throw / position mapping.
//
// Go: NewGermanUnpairedBracketsRule + de-DE JLanguageTool.GetAnalyzedSentence (real analysis).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// TestGenericUnpairedBracketsRule_GermanRule ports de.GenericUnpairedBracketsRuleTest.testGermanRule.
func TestGenericUnpairedBracketsRule_GermanRule(t *testing.T) {
	// Java: getBracketsRule finds GermanUnpairedBracketsRule on de-DE; construct directly.
	rule := NewGermanUnpairedBracketsRule(nil)
	lt := languagetool.NewJLanguageTool("de-DE")

	assertMatches := func(input string, expected int) {
		t.Helper()
		// Java: rule.match(Collections.singletonList(lt.getAnalyzedSentence(input)))
		sent := lt.GetAnalyzedSentence(input)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", input)
		n := len(rule.MatchList([]*languagetool.AnalyzedSentence{sent}))
		require.Equal(t, expected, n, "input=%q", input)
	}

	// correct sentences (Java assertMatches … 0):
	assertMatches("(Das sind die Sätze, die sie testen sollen).", 0)
	assertMatches("(Das sind die {Sätze}, die sie testen sollen).", 0)
	assertMatches("(Das sind die [Sätze], die sie testen sollen).", 0)
	assertMatches("(Das sind die Sätze (noch mehr Klammern [schon wieder!]), die sie testen sollen).", 0)
	assertMatches("Das ist ein Satz mit Smiley :-)", 0)
	assertMatches("Das ist auch ein Satz mit Smiley ;-)", 0)
	assertMatches("Das ist ein Satz mit Smiley :)", 0)
	assertMatches("Das ist ein Satz mit Smiley :(", 0)
	assertMatches("Die URL lautet https://de.wikipedia.org/wiki/Schlammersdorf_(Adelsgeschlecht)", 0)
	assertMatches("Die URL lautet https://de.wikipedia.org/wiki/Schlammersdorf_(Adelsgeschlecht).", 0)
	assertMatches("(Die URL lautet https://de.wikipedia.org/wiki/Schlammersdorf_(Adelsgeschlecht))", 0)
	assertMatches("(Die URL lautet https://de.wikipedia.org/wiki/Schlammersdorf)", 0)
	assertMatches("(Die URL lautet https://de.wikipedia.org/wiki/Schlammersdorf oder so)", 0)
	assertMatches("(Die URL lautet: http://www.pariscinema.org/).", 0)

	// Meta: id / URL / match.Rule attachment (Java setUrl on this DE rule; getId UNPAIRED_BRACKETS)
	require.Equal(t, "UNPAIRED_BRACKETS", rule.GetID())
	require.Equal(t, "https://languagetool.org/insights/de/beitrag/klammern/", rule.GetURL())
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{lt.GetAnalyzedSentence("Auch)")})
	require.NotEmpty(t, ms)
	require.Equal(t, rule, ms[0].GetRule())

	// Example pair from Java constructor
	require.Len(t, rule.GetIncorrectExamples(), 1)
	require.Len(t, rule.GetCorrectExamples(), 1)
	require.Equal(t, []string{"("}, rule.GetIncorrectExamples()[0].GetCorrections())

	// Start/end symbols (Java DE_START_SYMBOLS / DE_END_SYMBOLS; quotes moved to GermanUnpairedQuotesRule)
	require.Equal(t, []string{"[", "(", "{"}, rule.StartSymbols)
	require.Equal(t, []string{"]", ")", "}"}, rule.EndSymbols)

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
