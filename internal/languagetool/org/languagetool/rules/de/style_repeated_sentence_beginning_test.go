package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func artNomSent(art, noun, restVerb string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS(art, "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS(noun, "SUB:NOM:SIN:NEU", noun),
		atrWithPOS(restVerb, "VER:3:SIN:PRT:NON", restVerb),
		atrWithPOS(".", "PKT", "."),
	))
}

func TestStyleRepeatedSentenceBeginning(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// Java example: three sentences starting with ART:DEF:NOM
	sents := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
		artNomSent("Die", "Reifen", "quietschten"),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 3, len(matches))

	// mixed starts — no streak of 3 (middle not ART/PRO:PER:NOM)
	sents2 := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Langsam", "ADV", "langsam"),
			atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen"),
			atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
			atrWithPOS("Hund", "SUB:NOM:SIN:MAS", "Hund"),
			atrWithPOS(".", "PKT", "."),
		)),
		artNomSent("Die", "Reifen", "quietschten"),
	}
	require.Equal(t, 0, len(rule.MatchList(sents2)))

	// untagged AnalyzePlain must not invent article matches
	plain := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund lief langsam über die Straße."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	require.Equal(t, 0, len(rule.MatchList(plain)))
}
