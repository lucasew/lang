package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNeedsToBePlural(t *testing.T) {
	require.True(t, needsToBePlural("frau"))
	require.True(t, needsToBePlural("experte"))
	require.False(t, needsToBePlural("haus"))
	require.False(t, needsToBePlural(""))
}

func TestProcessTwoPart_PluralOnlyRejected(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	// no dict: IsMisspelled false for everything when FilterDict unavailable —
	// set override so part2 is accepted, part1 path uses TagPOS only
	r := NewGermanSpellerRule(nil)
	r.IsMisspelledOverride = func(w string) bool {
		// only flag nonsense; accept Diagramm / plural forms as spelled OK
		return w == "xyz"
	}
	// pure plural part1 "Frauen" (PLU only), part2 "Diagramm" not subVerInf → reject
	r.TagPOS = func(w string) []string {
		switch w {
		case "Frauen":
			return []string{"SUB:NOM:PLU:FEM"}
		case "Diagramm":
			return []string{"SUB:NOM:SIN:NEU"}
		case "Frau":
			return []string{"SUB:NOM:SIN:FEM"}
		default:
			return nil
		}
	}
	r.LemmaOf = func(w string) string {
		if w == "Frauen" || w == "Frau" {
			return "Frau"
		}
		return ""
	}
	// Frau is needsToBePlural — singular form of part1 rejected when lemma needs plural
	// part1WithoutInfixS = Frauen is PLU only → first arm: needsToBePlural(frau)=true so first reject skipped
	// second arm: needsToBePlural && isNounNomSin(Frauen)? Frauen is not SIN → second arm false
	// so may fall through to main arms — Frauen isNounNom? SUB:NOM:PLU matches isNounNom (SUB:NOM prefix)
	// Actually isNounNom checks SUB:NOM prefix — PLU tags match. Main arm without trailing s:
	// isNounNom(Frauen) true, !needsInfixS → would ACCEPT. That's different from pure "plural only reject".

	// Use lemma that is NOT needsToBePlural so pure-plural reject fires:
	r.LemmaOf = func(w string) string {
		if w == "Häuser" || w == "Haus" {
			return "Haus"
		}
		return ""
	}
	r.TagPOS = func(w string) []string {
		switch w {
		case "Häuser":
			return []string{"SUB:NOM:PLU:NEU"} // PLU only
		case "Diagramm":
			return []string{"SUB:NOM:SIN:NEU"}
		default:
			return nil
		}
	}
	require.False(t, r.ProcessTwoPartCompounds("Häuser", "Diagramm"),
		"pure plural part1 without special cases must be rejected")

	// needsToBePlural lemma with singular surface → reject
	r.TagPOS = func(w string) []string {
		switch w {
		case "Frau":
			return []string{"SUB:NOM:SIN:FEM"}
		case "Diagramm":
			return []string{"SUB:NOM:SIN:NEU"}
		default:
			return nil
		}
	}
	r.LemmaOf = func(w string) string {
		if w == "Frau" {
			return "Frau"
		}
		return ""
	}
	require.False(t, r.ProcessTwoPartCompounds("Frau", "Diagramm"),
		"needsToBePlural lemma with SIN surface must be rejected")
}
