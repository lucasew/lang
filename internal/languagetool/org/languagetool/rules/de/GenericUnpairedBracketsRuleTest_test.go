package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GenericUnpairedBracketsRuleTest.java
// Uses GermanUnpairedBracketsRule (brackets only, matching current DE Java symbols).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGenericUnpairedBracketsRule_GermanRule(t *testing.T) {
	rule := NewGermanUnpairedBracketsRule(nil)
	require.Equal(t, "UNPAIRED_BRACKETS", rule.GetID())
	require.Contains(t, rule.GetURL(), "klammern")

	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	// correct sentences (Java assertMatches … 0)
	require.Equal(t, 0, matchN("(Das sind die Sätze, die sie testen sollen)."))
	require.Equal(t, 0, matchN("(Das sind die {Sätze}, die sie testen sollen)."))
	require.Equal(t, 0, matchN("(Das sind die [Sätze], die sie testen sollen)."))
	require.Equal(t, 0, matchN("(Das sind die Sätze (noch mehr Klammern [schon wieder!]), die sie testen sollen)."))
	// smileys (Java isNoException)
	require.Equal(t, 0, matchN("Das ist ein Satz mit Smiley :-)"))
	require.Equal(t, 0, matchN("Das ist auch ein Satz mit Smiley ;-)"))
	require.Equal(t, 0, matchN("Das ist ein Satz mit Smiley :)"))
	require.Equal(t, 0, matchN("Das ist ein Satz mit Smiley :("))

	// URL with paren as single previous token (Java isNoException https?://.+ containing '(')
	// AnalyzePlain may split URLs; inject one token like Java WordTokenizer keeps it.
	urlSent := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART", "die"),
		atrWithPOS("URL", "SUB", "URL"),
		atrWithPOS("lautet", "VER", "lauten"),
		atrWithPOS("https://de.wikipedia.org/wiki/Schlammersdorf_(Adelsgeschlecht)", "URL", "https://…"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{urlSent})))

	// outer parens around URL sentence (balanced)
	require.Equal(t, 0, matchN("(Die URL lautet: http://www.pariscinema.org/)."))

	// unpaired
	require.Equal(t, 1, matchN("Auch)"))
	require.Equal(t, 1, matchN("Dem Präsidenten des Deutschen Bauernverbands (DBV zufolge habe die Dürre einen Schaden verursacht."))

	// soft hyphen neighbors (Java: used to map wrong positions; must not panic)
	for _, s := range []string{
		"Im Kran\u00ADken\u00ADhaus. Auch)",
		"Ein Kran\u00ADken\u00ADhaus. Auch)",
		"Das Kran\u00ADken\u00ADhaus. Auch)",
		"Kran\u00ADken\u00ADhaus. Auch)",
		"Kran\u00ADken\u00ADhaus. (Auch",
	} {
		_ = rule.MatchList(languagetool.AnalyzeTextLocal(s))
	}

	ms := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Auch)")})
	require.NotEmpty(t, ms)
	require.Equal(t, rule, ms[0].GetRule())
}
