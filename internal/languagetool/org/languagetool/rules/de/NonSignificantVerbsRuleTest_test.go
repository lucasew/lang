package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/NonSignificantVerbsRuleTest.java
// Java uses tagged analysis (hasAnyLemma); inject lemmas here.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of NonSignificantVerbsRuleTest.testRule (lemma-gated).
func TestNonSignificantVerbsRule_Rule(t *testing.T) {
	r := NewNonSignificantVerbsRule(nil)
	// machen + tun in s1; sein in s2 (Java multi-sentence expects 3)
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wenn", "KON:UNT", "wenn"),
		atrWithPOS("man", "PRO:IND:NOM:SIN:MAS", "man"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS("kann", "VER:MOD:3:SIN:PRS:SFT", "können"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("sollte", "VER:MOD:3:SIN:PRT:SFT", "sollen"),
		atrWithPOS("man", "PRO:IND:NOM:SIN:MAS", "man"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("tun", "VER:INF:NON", "tun"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "PRO:DEM:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("so", "ADV", "so"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 2, len(r.Match(s1))) // machen, tun
	require.Equal(t, 1, len(r.Match(s2))) // ist/sein

	// "Der Vorgang war abgeschlossen." — sein + PA2 exception → 0
	abgeschlossen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Vorgang", "SUB:NOM:SIN:MAS", "Vorgang"),
		atrWithPOS("war", "VER:3:SIN:PRT:SFT", "sein"),
		atrWithPOS("abgeschlossen", "VER:PA2:SFT", "abschließen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Empty(t, r.Match(abgeschlossen))

	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Vorgang", "SUB:NOM:SIN:MAS", "Vorgang"),
		atrWithPOS("endete", "VER:3:SIN:PRT:SFT", "enden"),
		atrWithPOS("plötzlich", "ADV", "plötzlich"),
		atrWithPOS(".", "PKT", "."),
	))))

	// untagged must not invent
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Er machte einen Kuchen.")))
}
