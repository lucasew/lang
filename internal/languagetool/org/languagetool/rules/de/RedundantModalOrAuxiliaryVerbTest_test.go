package de

// Twin of RedundantModalOrAuxiliaryVerbTest — Java uses tagged analysis (VER:MOD/VER:AUX).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRedundantModalOrAuxiliaryVerb_Rule(t *testing.T) {
	rule := NewRedundantModalOrAuxiliaryVerb(nil)

	// "… werde ich … und … werde ich …" — VER:AUX after und
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
	require.Equal(t, 1, len(rule.Match(werde)))

	// "… hat … und hat …"
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

	// "… ist gut und … ist gut"
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

	// no modal/aux repeat — untagged must not invent hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tom kauft Äpfel und Mary isst Bananen."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Erst werde ich die Preise vergleichen und erst dann werde ich entscheiden."))))
}
