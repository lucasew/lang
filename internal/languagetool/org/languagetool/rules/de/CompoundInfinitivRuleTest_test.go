package de

// Twin of CompoundInfinitivRuleTest — Java uses ZUS + VER:INF + speller.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundInfinitivRule_Rule(t *testing.T) {
	// Speller: joined particle+infinitive is a known compound (Java !isMisspelled)
	known := map[string]struct{}{
		"saubermachen": {},
		"vorbeikommen": {},
	}
	rule := NewCompoundInfinitivRule(nil)
	rule.IsMisspelled = func(w string) bool {
		_, ok := known[w]
		return !ok
	}

	// "sauber zu machen" — ZUS + zu + VER:INF
	sauber := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("brachte", "VER:3:SIN:PRT:SFT", "bringen"),
		atrWithPOS("ihn", "PRO:PER:AKK:SIN:MAS", "er"),
		atrWithPOS("dazu", "ADV", "dazu"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("mein", "PRO:POS:AKK:SIN:NEU", "mein"),
		atrWithPOS("Zimmer", "SUB:AKK:SIN:NEU", "Zimmer"),
		atrWithPOS("sauber", "ZUS", "sauber"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(sauber)))

	vorbei := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:MAS", "du"),
		atrWithPOS("brauchst", "VER:2:SIN:PRS:SFT", "brauchen"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("bei", "PRP:DAT", "bei"),
		atrWithPOS("mir", "PRO:PER:DAT:SIN:MAS", "ich"),
		atrWithPOS("vorbei", "ZUS", "vorbei"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("kommen", "VER:INF:NON", "kommen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(vorbei)))

	// "Fang an zu zählen" — anti-pattern / not ZUS particle path for "an" with Fang exception
	// Without immunization, "an" as ZUS + zählen INF may still hit if speller knows "anzählen"
	// Java test expects 0 — anti-pattern covers fang|… + ADV + an + zu
	fang := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Fang", "VER:IMP:SIN:SFT", "fangen"),
		atrWithPOS("an", "ZUS", "an"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("zählen", "VER:INF:NON", "zählen"),
		atrWithPOS(".", "PKT", "."),
	))
	// isException: VER:IMP lemma + "an"+"fangen" spellcheck — mark known
	rule.IsMisspelled = func(w string) bool {
		if w == "anfangen" || w == "Anfangen" {
			return false
		}
		_, ok := known[w]
		return !ok
	}
	require.Equal(t, 0, len(rule.Match(fang)))

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Du brauchst nicht bei mir vorbei zu kommen."))))
}
