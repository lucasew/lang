package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestAgreementRule2_MorphMismatch(t *testing.T) {
	// Kleiner (MAS) Haus (NEU) at sentence start with tags
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Kleiner", "ADJ:NOM:SIN:MAS:GRU:SOL", "klein"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule2(nil).Match(sent)
	require.NotEmpty(t, ms)
	require.Equal(t, agreement2Short, ms[0].ShortMessage)
}

func TestAgreementRule2_MorphOK(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Kleines", "ADJ:NOM:SIN:NEU:GRU:SOL", "klein"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule2(nil).Match(sent)
	require.Empty(t, ms)
}

func TestAgreementRule2_AntiPatternImmunizesWillkommen(t *testing.T) {
	// Willkommen + SUB is an anti-pattern
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Willkommen", "ADJ:NOM:SIN:NEU:GRU:SOL", "willkommen"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	// positions for matcher
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		if n := len(t.GetToken()); n > 0 {
			pos += n + 1
		}
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule2(nil)
	imm := r.getSentenceWithImmunization(sent)
	any := false
	for _, tok := range imm.GetTokensWithoutWhitespace() {
		if tok.IsImmunized() {
			any = true
		}
	}
	require.True(t, any)
	require.Empty(t, r.Match(sent))
}

func TestAgreementRule2_SuggestionsWithSynth(t *testing.T) {
	// Java: synthesize(adjToken, tag, true) → SynthesizeRE(..., true)
	synth := synthesis.FuncSynthesizer{
		SynthRE: func(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
			require.True(t, re)
			if posTag == "ADJ:NOM:SIN:NEU:GRU:SOL" {
				return []string{"kleines"}, nil
			}
			return nil, nil
		},
	}
	ss := languagetool.SentenceStartTagName
	toks := withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Kleiner", "ADJ:NOM:SIN:MAS:GRU:SOL", "klein"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule2(nil).WithSynth(synth).Match(sent)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "Kleines Haus")
}

func TestAgreementRule2_EmptyCategoriesIsMismatch(t *testing.T) {
	// Java retainCommonCategories: empty set1 retainAll set2 → empty → RuleMatch.
	// ADJ:PRD:GRU yields no casus/numerus/genus categories (e.g. "Wirtschaftlich Wachstum").
	ss := languagetool.SentenceStartTagName
	toks := withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Wirtschaftlich", "ADJ:PRD:GRU", "wirtschaftlich"),
		atrWithPOS("Wachstum", "SUB:NOM:SIN:NEU", "Wachstum"),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule2(nil).Match(sent)
	require.Equal(t, 1, len(ms))
	require.Equal(t, agreement2Short, ms[0].ShortMessage)
}

func TestAgreementRule2_ImmunizedContinueNotBreak(t *testing.T) {
	// Java: immunized/unter → continue (not break). "unter" then non-ADJ breaks.
	ss := languagetool.SentenceStartTagName
	toks := withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Unter", "ADJ:PRD:GRU", "unter"),
		atrWithPOS("Berücksichtigung", "SUB:DAT:SIN:FEM", "Berücksichtigung"),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	require.Empty(t, NewAgreementRule2(nil).Match(sent))
}
