package de

// Twin of InsertCommaFilterTest — POS-dependent branches need a tagger (Java GermanTagger).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertCommaFilter_Filter(t *testing.T) {
	f := NewInsertCommaFilter()
	// two tokens: no POS required
	require.Equal(t, []string{"hoffe, es"}, f.Suggest("hoffe es"))

	// three tokens without tagger: fail-closed (no invent both placements)
	require.Empty(t, f.Suggest("Ich hoffe es"))

	// with POS twin of Java hasTag checks
	f.TagToken = func(w string) []string {
		switch w {
		case "hoffe":
			return []string{"VER:1:SIN:PRÄ:NON"}
		case "es":
			return []string{"PRO:PER:NOM:SIN:3:NEU"}
		case "geht":
			return []string{"VER:3:SIN:PRÄ:SFT"}
		case "Sag", "sag":
			return []string{"VER:IMP:SIN:SFT"}
		case "mal":
			return []string{"ADV:TMP"}
		case "hast":
			return []string{"VER:2:SIN:PRÄ:NON"}
		case "denke":
			return []string{"VER:1:SIN:PRÄ:NON"}
		case "hier":
			return []string{"ADV:LOK"}
		case "kann":
			return []string{"VER:3:SIN:PRÄ:NON"}
		}
		return nil
	}
	// "hoffe es geht" → VER + PRO:PER → "hoffe, es geht"
	require.Equal(t, []string{"hoffe, es geht"}, f.Suggest("hoffe es geht"))
	// "Sag mal hast" → SAGT + mal + VER
	require.Equal(t, []string{"Sag mal, hast"}, f.Suggest("Sag mal hast"))
	// "denke hier kann" → VER + ADV + VER
	require.Equal(t, []string{"denke, hier kann"}, f.Suggest("denke hier kann"))
}
