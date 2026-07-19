package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WiederVsWiderRuleTest.java
// Java uses getAnalyzedSentence (lemma "spiegeln"); inject lemmas here (no surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWiederVsWiderRule_Rule(t *testing.T) {
	rule := NewWiederVsWiderRule(nil)
	// tag: token with lemma spiegeln on finite forms; plain tokens otherwise
	sent := func(parts ...string) *languagetool.AnalyzedSentence {
		toks := []*languagetool.AnalyzedTokenReadings{sentStartATR()}
		for _, p := range parts {
			switch p {
			case "spiegelt", "spiegeln", "spiegelte":
				toks = append(toks, atrWithPOS(p, "VER:3:SIN:PRS:SFT", "spiegeln"))
			default:
				// untagged surface is fine for non-lemma checks
				toks = append(toks, atrWithPOS(p, "UNKNOWN", p))
			}
		}
		return languagetool.NewAnalyzedSentence(withPositions(toks...))
	}
	ok := func(parts ...string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(sent(parts...))), parts)
	}
	bad := func(parts ...string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(sent(parts...))), parts)
	}
	ok("Das", "spiegelt", "wider", ",", "wie", "es", "wieder", "läuft", ".")
	ok("Das", "spiegelt", "die", "Situation", "gut", "wider", ".")
	ok("Das", "spiegelt", "die", "Situation", ".")
	ok("Immer", "wieder", "spiegelt", "das", "die", "Situation", ".")
	ok("Immer", "wieder", "spiegelt", "das", "die", "Situation", "wider", ".")
	ok("Das", "spiegelt", "wieder", "wider", ",", "wie", "es", "läuft", ".")

	bad("Das", "spiegelt", "wieder", ",", "wie", "es", "wieder", "läuft", ".")
	bad("Sie", "spiegeln", "das", "Wachstum", "der", "Stadt", "wieder", ".")
	bad("Das", "spiegelt", "die", "Situation", "gut", "wieder", ".")
	bad("Immer", "wieder", "spiegelt", "das", "die", "Situation", "wieder", ".")
	bad("Immer", "wieder", "spiegelte", "das", "die", "Situation", "wieder", ".")
}
