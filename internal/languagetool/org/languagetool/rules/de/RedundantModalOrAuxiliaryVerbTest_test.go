package de

// Twin of RedundantModalOrAuxiliaryVerbTest — Java uses tagged analysis (VER:MOD/VER:AUX).
// Morph/POS inject only; untagged AnalyzePlain remains fail-closed.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRedundantModalOrAuxiliaryVerb_Rule(t *testing.T) {
	rule := NewRedundantModalOrAuxiliaryVerb(nil)

	// Java match: "Erst werde ich … und erst dann werde ich …"
	werde := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Erst", "ADV", "erst"),
		atrWithPOS("werde", "VER:AUX:1:SIN:PRS:SFT", "werden"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("die", "ART:DEF:AKK:PLU:MAS", "der"),
		atrWithPOS("Preise", "SUB:AKK:PLU:MAS", "Preis"),
		atrWithPOS("vergleichen", "VER:INF:NON", "vergleichen"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("erst", "ADV", "erst"),
		atrWithPOS("dann", "ADV", "dann"),
		atrWithPOS("werde", "VER:AUX:1:SIN:PRS:SFT", "werden"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("entscheiden", "VER:INF:NON", "entscheiden"),
		atrWithPOS(".", "PKT", "."),
	))
	ms := rule.Match(werde)
	require.Equal(t, 1, len(ms))
	// "werde ich" repeats as multi-token → Satzteil message (not single Hilfsverb)
	require.Contains(t, ms[0].GetMessage(), "redundant")
	require.Empty(t, ms[0].GetShortMessage())

	// Java: "Sie hat das Foto … angeschaut und hat gelacht."
	hat := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("hat", "VER:AUX:3:SIN:PRS:SFT", "haben"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Foto", "SUB:AKK:SIN:NEU", "Foto"),
		atrWithPOS("angeschaut", "VER:PA2:NON", "anschauen"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("hat", "VER:AUX:3:SIN:PRS:SFT", "haben"),
		atrWithPOS("gelacht", "VER:PA2:SFT", "lachen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(hat)))

	// Java: "Das Essen ist gut und der Service hier ist gut."
	ist := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Essen", "SUB:NOM:SIN:NEU", "Essen"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("gut", "ADJ:PRD:GRU", "gut"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Service", "SUB:NOM:SIN:MAS", "Service"),
		atrWithPOS("hier", "ADV", "hier"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("gut", "ADJ:PRD:GRU", "gut"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(ist)))

	// Java: "Ich bin gern in eurer Mitte und ich bin gern zu Gast bei euch."
	// second "bin" after "ich" (PRO:PER before repeated AUX) → Satzteil path
	bin := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("bin", "VER:AUX:1:SIN:PRS:SFT", "sein"),
		atrWithPOS("gern", "ADV", "gern"),
		atrWithPOS("in", "APPR:DAT", "in"),
		atrWithPOS("eurer", "PRO:POS:DAT:SIN:FEM", "euer"),
		atrWithPOS("Mitte", "SUB:DAT:SIN:FEM", "Mitte"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("bin", "VER:AUX:1:SIN:PRS:SFT", "sein"),
		atrWithPOS("gern", "ADV", "gern"),
		atrWithPOS("zu", "APPR:DAT", "zu"),
		atrWithPOS("Gast", "SUB:DAT:SIN:MAS", "Gast"),
		atrWithPOS("bei", "APPR:DAT", "bei"),
		atrWithPOS("euch", "PRO:PER:DAT:PLU:*", "ihr"),
		atrWithPOS(".", "PKT", "."),
	))
	ms = rule.Match(bin)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "Satzteil")

	// Java: "Wann ist jemand kühn und wann ist jemand tollkühn?"
	// ist ... jemand und ... ist jemand → sub text on repeated clause
	wann := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wann", "PWAV", "wann"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("jemand", "PIS:NOM:SIN:MAS", "jemand"),
		atrWithPOS("kühn", "ADJ:PRD:GRU", "kühn"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("wann", "PWAV", "wann"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("jemand", "PIS:NOM:SIN:MAS", "jemand"),
		atrWithPOS("tollkühn", "ADJ:PRD:GRU", "tollkühn"),
		atrWithPOS("?", "PKT", "?"),
	))
	// jemand is PIS not PRO:IND in all taggers; Java uses PRO:IND for jemand?
	// If PRO:IND: inject as PRO:IND for the branch that matches next-token equality
	// Re-tag jemand as PRO:IND to match Java morph class for indefinite pronouns.
	// Actually HasPosTagStartingWith("PRO:IND") — PIS may not match. Use PRO:IND:
	wann = languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wann", "PWAV", "wann"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("jemand", "PRO:IND:NOM:SIN:MAS", "jemand"),
		atrWithPOS("kühn", "ADJ:PRD:GRU", "kühn"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("wann", "PWAV", "wann"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("jemand", "PRO:IND:NOM:SIN:MAS", "jemand"),
		atrWithPOS("tollkühn", "ADJ:PRD:GRU", "tollkühn"),
		atrWithPOS("?", "PKT", "?"),
	))
	require.Equal(t, 1, len(rule.Match(wann)))

	// Java modal: "Wir müssen wissen, was wir tun sollen und wie wir es tun sollen."
	// comma breaks first scan; second sollen after und
	sollen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wir", "PRO:PER:NOM:PLU:*", "wir"),
		atrWithPOS("müssen", "VER:MOD:1:PLU:PRS:SFT", "müssen"),
		atrWithPOS("wissen", "VER:INF:NON", "wissen"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("was", "PWS:AKK:SIN:NEU", "was"),
		atrWithPOS("wir", "PRO:PER:NOM:PLU:*", "wir"),
		atrWithPOS("tun", "VER:INF:NON", "tun"),
		atrWithPOS("sollen", "VER:MOD:1:PLU:PRS:SFT", "sollen"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("wie", "PWAV", "wie"),
		atrWithPOS("wir", "PRO:PER:NOM:PLU:*", "wir"),
		atrWithPOS("es", "PRO:PER:AKK:SIN:NEU", "es"),
		atrWithPOS("tun", "VER:INF:NON", "tun"),
		atrWithPOS("sollen", "VER:MOD:1:PLU:PRS:SFT", "sollen"),
		atrWithPOS(".", "PKT", "."),
	))
	// Note: comma breaks after müssen — the sollen...und...sollen is the match region
	require.Equal(t, 1, len(rule.Match(sollen)))

	// Java: Tom ist um halb drei angekommen und Mary ist kurze Zeit später angekommen.
	// participle path (PA2 before und, same participle after second AUX)
	angekommen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Tom", "EIG:NOM:SIN:MAS", "Tom"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("um", "APPR:AKK", "um"),
		atrWithPOS("halb", "ADV", "halb"),
		atrWithPOS("drei", "CARD", "drei"),
		atrWithPOS("angekommen", "PA2", "ankommen"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("Mary", "EIG:NOM:SIN:FEM", "Mary"),
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("kurze", "ADJA:AKK:SIN:FEM:GRU:IND", "kurz"),
		atrWithPOS("Zeit", "SUB:AKK:SIN:FEM", "Zeit"),
		atrWithPOS("später", "ADV", "später"),
		atrWithPOS("angekommen", "PA2", "ankommen"),
		atrWithPOS(".", "PKT", "."),
	))
	// hasParticipleAt looks for PA2 *before* conjunction: tokens[nConjunction-1] must be PA2
	// Here token before und is "angekommen" ✓; after second ist find angekommen at end
	ms = rule.Match(angekommen)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "Satzteil")

	// no match — untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tom kauft Äpfel und Mary isst Bananen."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Erst werde ich die Preise vergleichen und erst dann werde ich entscheiden."))))

	// Java no-match: Johnny hat Alice vorgeschlagen und sie hat akzeptiert.
	// PRO:PER "sie" before second hat → shouldSkipRedundantBranch (PRO:PER on nt-1).
	johnny := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Johnny", "EIG:NOM:SIN:MAS", "Johnny"),
		atrWithPOS("hat", "VER:AUX:3:SIN:PRS:SFT", "haben"),
		atrWithPOS("Alice", "EIG:AKK:SIN:FEM", "Alice"),
		atrWithPOS("vorgeschlagen", "PA2", "vorschlagen"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("hat", "VER:AUX:3:SIN:PRS:SFT", "haben"),
		atrWithPOS("akzeptiert", "PA2", "akzeptieren"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(johnny)))

	// Java no-match: "Sie mag niemanden und niemand mag sie." — mag not VER:AUX/MOD → 0
	// With only non-MOD/AUX tags:
	mag := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("mag", "VER:1:SIN:PRS:SFT", "mögen"), // not VER:MOD in this fixture
		atrWithPOS("niemanden", "PIS:AKK:SIN:MAS", "niemand"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("niemand", "PIS:NOM:SIN:MAS", "niemand"),
		atrWithPOS("mag", "VER:1:SIN:PRS:SFT", "mögen"),
		atrWithPOS("sie", "PRO:PER:AKK:SIN:FEM", "sie"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(mag)))
}

func TestRedundantModalOrAuxiliaryVerb_Meta(t *testing.T) {
	r := NewRedundantModalOrAuxiliaryVerb(nil)
	require.Equal(t, "REDUNDANT_MODAL_VERB", r.GetID())
	require.Equal(t, "Redundantes Modal- oder Hilfsverb", r.GetDescription())
	require.True(t, r.IsDefaultOff())
	require.NotNil(t, r.GetCategory())
}
