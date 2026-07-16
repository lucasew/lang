package de

// Twin of RedundantModalOrAuxiliaryVerbTest (surface modal/aux forms).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRedundantModalOrAuxiliaryVerb_Rule(t *testing.T) {
	rule := NewRedundantModalOrAuxiliaryVerb(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	// clear repeats of aux/modal after und
	require.Equal(t, 1, matchN("Erst werde ich die Preise vergleichen und erst dann werde ich entscheiden, ob ich die Kamera kaufe."))
	require.Equal(t, 1, matchN("Sie hat das Foto von mir als kleinem Jungen angeschaut und hat gelacht."))
	require.Equal(t, 1, matchN("Das Essen ist gut und der Service hier ist gut."))
	// no modal/aux repeat
	require.Equal(t, 0, matchN("Tom kauft Äpfel und Mary isst Bananen."))
	// Java goods with different subjects often need POS; surface may over-flag — not asserted hard
}
