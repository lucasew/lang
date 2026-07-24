package de

// Twin of StyleRepeatedSentenceBeginning (Java MIN_REPEATED=3; ART:DEF:NOM / PRO:PER:NOM).
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

func proNomSent(pro, restVerb string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS(pro, "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS(restVerb, "VER:3:SIN:PRT:NON", restVerb),
		atrWithPOS(".", "PKT", "."),
	))
}

func TestStyleRepeatedSentenceBeginning(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	require.Equal(t, "STYLE_REPEATED_SENTENCE_BEGINNING", rule.GetID())
	require.Equal(t, "Subjekt als wiederholter Satzanfang", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 3, rule.MinRepeated)
	require.Equal(t, 3, rule.MinToCheckParagraph())
	require.NotEmpty(t, rule.GetIncorrectExamples())

	// Java example: three sentences starting with ART:DEF:NOM → marker ART…SUB
	sents := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
		artNomSent("Die", "Reifen", "quietschten"),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 3, len(matches))
	for _, m := range matches {
		require.Equal(t, "Subjekt als wiederholter Satzanfang", m.GetMessage())
	}
	// first match local (pos=0): ART tokens[1] … SUB tokens[2] (Java endPos after SUB)
	toks0 := sents[0].GetTokensWithoutWhitespace()
	require.Equal(t, toks0[1].GetStartPos(), matches[0].GetFromPos())
	require.Equal(t, toks0[2].GetEndPos(), matches[0].GetToPos())

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

	// PRO:PER:NOM ×3 → 3 (span is pronoun only)
	proSents := []*languagetool.AnalyzedSentence{
		proNomSent("Er", "lief"),
		proNomSent("Sie", "rief"),
		proNomSent("Es", "ging"),
	}
	// retag pronouns with correct lemmas/surfaces for PRO:PER:NOM
	proSents = []*languagetool.AnalyzedSentence{
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
			atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen"),
			atrWithPOS(".", "PKT", "."),
		)),
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Sie", "PRO:PER:NOM:SIN:FEM", "sie"),
			atrWithPOS("rief", "VER:3:SIN:PRT:SFT", "rufen"),
			atrWithPOS(".", "PKT", "."),
		)),
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
			atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen"),
			atrWithPOS(".", "PKT", "."),
		)),
	}
	msPro := rule.MatchList(proSents)
	require.Equal(t, 3, len(msPro))
	// first span = pronoun only
	pt := proSents[0].GetTokensWithoutWhitespace()
	require.Equal(t, pt[1].GetStartPos(), msPro[0].GetFromPos())
	require.Equal(t, pt[1].GetEndPos(), msPro[0].GetToPos())

	// fewer than MIN_REPEATED → 0
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
	})))

	// untagged AnalyzePlain must not invent article matches
	plain := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund lief langsam über die Straße."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	require.Equal(t, 0, len(rule.MatchList(plain)))
}
