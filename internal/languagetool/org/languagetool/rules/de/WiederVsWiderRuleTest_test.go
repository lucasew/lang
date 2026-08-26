package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WiederVsWiderRuleTest.java
//
// Java: rule.match(lt.getAnalyzedSentence(text)).length == 0|1.
// Java match uses tokens[i].hasLemma("spiegeln") — real de-DE analysis must supply lemma
// "spiegeln" for finite forms; this twin injects that lemma (no surface invent).
// Untagged plain analysis must not fire (meta: no invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWiederVsWiderRule_Rule(t *testing.T) {
	// Java: new WiederVsWiderRule(TestTools.getMessages("de"))
	rule := NewWiederVsWiderRule(nil)

	// tag: lemma "spiegeln" on finite forms (Java hasLemma); plain tokens otherwise
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
		ms := rule.Match(sent(parts...))
		require.Equal(t, 1, len(ms), parts)
		// Java-visible outcomes beyond count: id, suggestion, messages, span on "wieder"
		require.Equal(t, "DE_WIEDER_VS_WIDER", rule.GetID())
		require.Equal(t, []string{"wider"}, ms[0].GetSuggestedReplacements(), parts)
		require.Equal(t, "'wider' in 'widerspiegeln' wird mit 'i' geschrieben", ms[0].GetShortMessage(), parts)
		require.Contains(t, ms[0].GetMessage(), "widerspiegeln", parts)
		require.Contains(t, ms[0].GetMessage(), "mit 'i' statt mit 'ie'", parts)
		// Java marks current token (the "wieder" that set foundWieder)
		require.Equal(t, "wider", ms[0].GetSuggestedReplacements()[0])
		span := ms[0].GetToPos() - ms[0].GetFromPos()
		require.Equal(t, len("wieder"), span, "match span is token.length of wieder: %v", parts)
	}

	// Java assertGood matrix:
	ok("Das", "spiegelt", "wider", ",", "wie", "es", "wieder", "läuft", ".")
	ok("Das", "spiegelt", "die", "Situation", "gut", "wider", ".")
	ok("Das", "spiegelt", "die", "Situation", ".")
	ok("Immer", "wieder", "spiegelt", "das", "die", "Situation", ".")
	ok("Immer", "wieder", "spiegelt", "das", "die", "Situation", "wider", ".")
	ok("Das", "spiegelt", "wieder", "wider", ",", "wie", "es", "läuft", ".")

	// Java assertBad matrix:
	bad("Das", "spiegelt", "wieder", ",", "wie", "es", "wieder", "läuft", ".")
	bad("Sie", "spiegeln", "das", "Wachstum", "der", "Stadt", "wieder", ".")
	bad("Das", "spiegelt", "die", "Situation", "gut", "wieder", ".")
	bad("Immer", "wieder", "spiegelt", "das", "die", "Situation", "wieder", ".")
	bad("Immer", "wieder", "spiegelte", "das", "die", "Situation", "wieder", ".")
}

func TestWiederVsWiderRule_Meta(t *testing.T) {
	r := NewWiederVsWiderRule(nil)
	require.Equal(t, "DE_WIEDER_VS_WIDER", r.GetID())
	require.Equal(t, "Möglicher Tippfehler 'spiegeln ... wieder (wider)'", r.GetDescription())
	require.Equal(t, 0, r.EstimateContextForSureMatch())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, "TYPOS", r.GetCategory().GetID().String())
	require.Len(t, r.GetIncorrectExamples(), 1)
	require.Len(t, r.GetCorrectExamples(), 1)
	// untagged must not invent (lemma spiegeln required — Java hasLemma only)
	require.Equal(t, 0, len(r.Match(languagetool.AnalyzePlain("Das spiegelt die Situation gut wieder."))))
	// look-ahead Java quirk preserved: wider at i+1 suppresses when length > i+2
	// "spiegelt wieder wider ." → tokens after SENT_START; at wieder, i+1=wider and length>i+2
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("spiegelt", "VER:3:SIN:PRS:SFT", "spiegeln"),
		atrWithPOS("wieder", "UNKNOWN", "wieder"),
		atrWithPOS("wider", "UNKNOWN", "wider"),
		atrWithPOS(".", "UNKNOWN", "."),
	)
	require.Equal(t, 0, len(r.Match(languagetool.NewAnalyzedSentence(toks))))
}
