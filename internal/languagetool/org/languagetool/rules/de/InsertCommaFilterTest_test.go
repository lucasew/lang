package de

// Twin of InsertCommaFilterTest — POS-dependent branches need a tagger (Java GermanTagger).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestInsertCommaFilter_Filter(t *testing.T) {
	f := NewInsertCommaFilter()
	// two tokens: no POS required
	require.Equal(t, []string{"hoffe, es"}, f.Suggest("hoffe es"))
	require.Equal(t, []string{"steht, parkt"}, f.Suggest("steht parkt"))

	// three tokens without tagger: fail-closed (no invent both placements)
	require.Empty(t, f.Suggest("Ich hoffe es"))

	// with POS twin of Java hasTag checks
	f.TagToken = func(w string) []string {
		switch w {
		case "hoffe", "Hoffe":
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
		case "ist":
			return []string{"VER:3:SIN:PRÄ:NON"}
		case "steht", "parkt":
			return []string{"VER:3:SIN:PRÄ:SFT"}
		}
		return nil
	}
	// "hoffe es geht" → VER + PRO:PER → "hoffe, es geht"
	require.Equal(t, []string{"hoffe, es geht"}, f.Suggest("hoffe es geht"))
	// "Sag mal hast" → SAGT + mal + VER
	require.Equal(t, []string{"Sag mal, hast"}, f.Suggest("Sag mal hast"))
	// "denke hier kann" → VER + ADV + VER
	require.Equal(t, []string{"denke, hier kann"}, f.Suggest("denke hier kann"))
	// "Hoffe bei euch ist" → patternTokenPos==1 BEI+DIR+VER (Java test case)
	require.Equal(t, []string{"Hoffe, bei euch ist"}, f.Suggest("Hoffe bei euch ist"))
}

// Twin of InsertCommaFilterTest.runFilter via AcceptRuleMatch
func TestInsertCommaFilter_AcceptRuleMatch_JavaCases(t *testing.T) {
	f := NewInsertCommaFilter()
	f.TagToken = func(w string) []string {
		switch w {
		case "steht", "parkt":
			return []string{"VER:3:SIN:PRÄ:SFT"}
		case "Hoffe", "hoffe":
			return []string{"VER:1:SIN:PRÄ:NON"}
		case "ist":
			return []string{"VER:3:SIN:PRÄ:NON"}
		}
		return nil
	}

	// Java: "steht parkt" → "[steht, parkt]"
	sent1 := languagetool.AnalyzePlain("Das Auto, das am Straßenrand steht parkt im Halteverbot.")
	m1 := rules.NewRuleMatch(rules.NewFakeRule("KOMMA"), sent1, 29, 40, "fake msg")
	m1.SetSuggestedReplacement("steht parkt")
	// patternTokenPos=7; atr from stands/parkt not needed for 2-part
	out1 := f.AcceptRuleMatch(m1, map[string]string{}, 7, nil, nil)
	require.NotNil(t, out1)
	require.Equal(t, []string{"steht, parkt"}, out1.GetSuggestedReplacements())

	// Java: "Hoffe bei euch ist" → "[Hoffe, bei euch ist]" with patternTokenPos=1
	sent2 := languagetool.AnalyzePlain("Hoffe bei euch ist alles gut.")
	m2 := rules.NewRuleMatch(rules.NewFakeRule("KOMMA"), sent2, 0, 7, "fake msg")
	m2.SetSuggestedReplacement("Hoffe bei euch ist")
	// pattern tokens only need length for some branches; BEI path uses parts only
	atr := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("Hoffe", "VER:1:SIN:PRÄ:NON", "hoffen"),
		atrWithPOS("bei", "APPR:DAT", "bei"),
		atrWithPOS("euch", "PRO:PER:DAT:PLU", "ihr"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
	}
	out2 := f.AcceptRuleMatch(m2, map[string]string{}, 1, atr, nil)
	require.NotNil(t, out2)
	require.Equal(t, []string{"Hoffe, bei euch ist"}, out2.GetSuggestedReplacements())
}
