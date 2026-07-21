package de

// Twin of CompoundInfinitivRuleTest — Java uses ZUS + VER:INF + speller.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundInfinitivRule_Rule(t *testing.T) {
	// Speller: joined particle+infinitive is a known compound (Java !isMisspelled)
	// Also "anfangen" etc. for exception scan.
	known := map[string]struct{}{
		"saubermachen": {},
		"vorbeikommen": {},
		"vorbeilassen": {},
		"anfangen":     {},
		"aufhören":     {}, // if scanned as particle+verb
	}
	rule := NewCompoundInfinitivRule(nil)
	rule.IsMisspelled = func(w string) bool {
		_, ok := known[w]
		return !ok
	}

	// Java: "Ich brachte ihn dazu, mein Zimmer sauber zu machen." → 1
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

	// Java: "Du brauchst nicht bei mir vorbei zu kommen." → 1
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

	// Java: "… die alte Dame vorbei zu lassen." → 1
	lassen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen"),
		atrWithPOS("zur", "APPRART:DAT:FEM", "zu"),
		atrWithPOS("Seite", "SUB:DAT:SIN:FEM", "Seite"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("um", "KOUI", "um"),
		atrWithPOS("die", "ART:DEF:AKK:SIN:FEM", "die"),
		atrWithPOS("alte", "ADJ:AKK:SIN:FEM:GRU:DEF", "alt"),
		atrWithPOS("Dame", "SUB:AKK:SIN:FEM", "Dame"),
		atrWithPOS("vorbei", "ZUS", "vorbei"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("lassen", "VER:INF:NON", "lassen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(lassen)))

	// Java: "Seine Frau gab vor zu schlafen." → 0 (isException: tokens[n-2] VER)
	// "gab" VER, "vor" ZUS, "zu", "schlafen"
	vor := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Seine", "PRO:POS:NOM:SIN:FEM", "sein"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS("gab", "VER:3:SIN:PRT:NON", "geben"),
		atrWithPOS("vor", "ZUS", "vor"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("schlafen", "VER:INF:NON", "schlafen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(vor)))

	// Java: "Mein Herz hörte auf zu schlagen." → 0 (VER before particle)
	auf := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Mein", "PRO:POS:NOM:SIN:NEU", "mein"),
		atrWithPOS("Herz", "SUB:NOM:SIN:NEU", "Herz"),
		atrWithPOS("hörte", "VER:3:SIN:PRT:SFT", "hören"),
		atrWithPOS("auf", "ZUS", "auf"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("schlagen", "VER:INF:NON", "schlagen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(auf)))

	// Java: "Fang an zu zählen." → 0 via isException verb scan (an+fangen known)
	fang := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Fang", "VER:IMP:SIN:SFT", "fangen"),
		atrWithPOS("an", "ZUS", "an"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("zählen", "VER:INF:NON", "zählen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(fang)))

	// Java: "Er hatte nichts weiter zu sagen" → 0 (weiter+sagen adj/special exception)
	// "weiter" is in ADJ_EXCEPTION? Looking Java: sagen+weiter exception
	weiter := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("hatte", "VER:3:SIN:PRT:NON", "haben"),
		atrWithPOS("nichts", "PIS", "nichts"),
		atrWithPOS("weiter", "ZUS", "weiter"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("sagen", "VER:INF:NON", "sagen"),
		atrWithPOS(".", "PKT", "."),
	))
	// n-2 is "nichts" not VER — but sagen+weiter exception
	require.Equal(t, 0, len(rule.Match(weiter)))

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Du brauchst nicht bei mir vorbei zu kommen."))))

	// Without speller, fail closed (all joins misspelled)
	noSpell := NewCompoundInfinitivRule(nil)
	require.Equal(t, 0, len(noSpell.Match(sauber)))
}

func TestCompoundInfinitivRule_IsPunctuationUTF16(t *testing.T) {
	require.True(t, isPunctuationCI("…"))
	require.True(t, isPunctuationCI("."))
	require.False(t, isPunctuationCI(".."))
	require.False(t, isPunctuationCI(""))
}
