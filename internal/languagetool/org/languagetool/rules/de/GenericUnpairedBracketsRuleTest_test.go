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
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	// correct sentences (from Java twin)
	require.Equal(t, 0, matchN("(Das sind die Sätze, die sie testen sollen)."))
	require.Equal(t, 0, matchN("(Das sind die {Sätze}, die sie testen sollen)."))
	require.Equal(t, 0, matchN("(Das sind die [Sätze], die sie testen sollen)."))
	require.Equal(t, 0, matchN("(Das sind die Sätze (noch mehr Klammern [schon wieder!]), die sie testen sollen)."))
	require.Equal(t, 0, matchN("Das ist ein Satz mit Smiley :-)"))
	require.Equal(t, 0, matchN("Das ist auch ein Satz mit Smiley ;-)"))
	// unpaired (Java example style: missing close paren)
	require.Equal(t, 1, matchN("Auch)"))
	require.Equal(t, 1, matchN("Dem Präsidenten des Deutschen Bauernverbands (DBV zufolge habe die Dürre einen Schaden verursacht."))
}
