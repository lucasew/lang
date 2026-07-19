package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestReplacePrepositionsByArticle_Ins(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("ins", "PRP:AKK:SIN", "in"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
	}
	m := replacePrepositionsByArticle(toks)
	require.Equal(t, ReplIns, m[0])
	require.Equal(t, "das", toks[0].GetToken())
	require.True(t, toks[0].HasPosTagStartingWith("ART:"))
	// noun unchanged
	require.Equal(t, "Haus", toks[1].GetToken())
}

func TestReplacePrepositionsByArticle_Zur(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("zur", "PRP:DAT:SIN", "zu"),
		atrWithPOS("Frau", "SUB:DAT:SIN:FEM", "Frau"),
	}
	m := replacePrepositionsByArticle(toks)
	require.Equal(t, ReplZur, m[0])
	require.Equal(t, "der", toks[0].GetToken())
}

func TestAgreementRule_InsWithWrongGender(t *testing.T) {
	// ins rewritten to das(NEU); die-tagged Haus? use FEM noun mismatch: das + Frau
	// "ins" + "Frau" → das AKK NEU vs Frau FEM → mismatch
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("ins", "PRP:AKK:SIN", "in"),
		atrWithPOS("Frau", "SUB:AKK:SIN:FEM", "Frau"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms, "ins Frau (→ das Frau) should mismatch")
}

func TestAgreementRule_InsWithNeutralNounOK(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("ins", "PRP:AKK:SIN", "in"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule(nil).Match(sent)
	for _, m := range ms {
		require.NotEqual(t, agreementShort, m.ShortMessage)
	}
}

func TestGetCompoundErrorDetNoun(t *testing.T) {
	// die + Original + Mail → compound suggestions when lt.check gate accepts both forms
	det := atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die")
	noun := atrWithPOS("Original", "SUB:NOM:SIN:NEU", "Original")
	next := atrWithPOS("Mail", "SUB:NOM:SIN:FEM", "Mail")
	// positions
	det.SetStartPos(0)
	noun.SetStartPos(4)
	next.SetStartPos(13)
	orig := []*languagetool.AnalyzedTokenReadings{det, noun, next}
	sent := languagetool.NewAnalyzedSentence(orig)
	ar := NewAgreementRule(nil)
	ar.CompoundPhraseValid = func(phrase string) bool { return true }
	rm := getCompoundErrorDetNoun(det, noun, 0, orig, sent, ar)
	require.NotNil(t, rm)
	require.Contains(t, rm.GetSuggestedReplacements(), "die Originalmail")
	require.Contains(t, rm.GetSuggestedReplacements(), "die Original-Mail")
}

func TestGetCompoundErrorDetNoun_FailClosedWithoutValid(t *testing.T) {
	// Without CompoundPhraseValid (Java lt.check), do not invent open-compound hits.
	det := atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die")
	noun := atrWithPOS("Original", "SUB:NOM:SIN:NEU", "Original")
	next := atrWithPOS("Mail", "SUB:NOM:SIN:FEM", "Mail")
	det.SetStartPos(0)
	noun.SetStartPos(4)
	next.SetStartPos(13)
	orig := []*languagetool.AnalyzedTokenReadings{det, noun, next}
	sent := languagetool.NewAnalyzedSentence(orig)
	rm := getCompoundErrorDetNoun(det, noun, 0, orig, sent, NewAgreementRule(nil))
	require.Nil(t, rm)
}

func TestAgreementRule_CompoundErrorOnMismatch(t *testing.T) {
	// Compound helper is covered by TestGetCompoundErrorDetNoun.
	// Match-level: DET–NOUN mismatch still fires when not immunized (two-token).
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Katalog", "SUB:NOM:SIN:NEU", "Katalog"),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms, "die Katalog should mismatch")
	require.Equal(t, agreementShort, ms[0].ShortMessage)
}

func TestApplyReplacementContractions_Ins(t *testing.T) {
	got := applyReplacementContractions([]string{"das Haus", "die Häuser", "dem Haus"}, ReplIns)
	require.Contains(t, got, "ins Haus")
	require.Contains(t, got, "in die Häuser")
	require.Contains(t, got, "im Haus")
}

func TestApplyReplacementContractions_Zur(t *testing.T) {
	got := applyReplacementContractions([]string{"der Frau", "dem Mann"}, ReplZur)
	require.Contains(t, got, "zur Frau")
	require.Contains(t, got, "zum Mann")
}

func TestAgreementSuggestor2_WithInsContraction(t *testing.T) {
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil {
				return nil, nil
			}
			if strings.Contains(posTag, "ART:") && strings.Contains(posTag, ":AKK:") && strings.Contains(posTag, "NEU") {
				return []string{"das"}, nil
			}
			if strings.Contains(posTag, "ART:") && strings.Contains(posTag, ":NOM:") && strings.Contains(posTag, "NEU") {
				return []string{"das"}, nil
			}
			if strings.Contains(posTag, "ART:") && strings.Contains(posTag, ":DAT:") {
				return []string{"dem"}, nil
			}
			if strings.Contains(posTag, "SUB:") {
				if strings.Contains(posTag, ":PLU:") {
					return []string{"Häuser"}, nil
				}
				return []string{"Haus"}, nil
			}
			return []string{token.GetToken()}, nil
		},
	}
	det := atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "das")
	noun := atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus")
	sugs := NewAgreementSuggestor2(synth, det, noun).WithReplacementType(ReplIns).GetSuggestions()
	require.NotEmpty(t, sugs)
	hasIns := false
	for _, s := range sugs {
		if strings.HasPrefix(s, "ins") || strings.HasPrefix(s, "im") || strings.HasPrefix(s, "in ") {
			hasIns = true
			break
		}
	}
	require.True(t, hasIns, "expected ins/im/in… contraction, got %v", sugs)
}
