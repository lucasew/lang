package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestAgreementSuggestor2(t *testing.T) {
	// Java skips edits==0 (identical phrase). Synth must return alternate forms.
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil {
				return nil, nil
			}
			// Tags: ART:DEF:AKK:SIN:MAS (no trailing colon after gender).
			if strings.Contains(posTag, "ART:") && strings.Contains(posTag, ":AKK:") {
				return []string{"den"}, nil
			}
			if strings.Contains(posTag, "ART:") && strings.Contains(posTag, ":NOM:") {
				return []string{"der"}, nil
			}
			if strings.Contains(posTag, "SUB:") {
				return []string{"Hund"}, nil
			}
			return []string{token.GetToken()}, nil
		},
	}
	det := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("der", strp("ART:DEF:NOM:SIN:MAS"), strp("der")), 0)
	noun := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Hund", strp("SUB:NOM:SIN:MAS"), strp("Hund")), 4)
	s := NewAgreementSuggestor2(synth, det, noun)
	got := s.GetSuggestions()
	require.NotEmpty(t, got)
	require.Contains(t, got, "den Hund")
}

func TestAgreementSuggestor2_FilterKeepsLowestEditTier(t *testing.T) {
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil {
				return nil, nil
			}
			if strings.Contains(posTag, "ART:") && strings.Contains(posTag, ":AKK:") {
				return []string{"den"}, nil
			}
			if strings.Contains(posTag, "ART:") {
				return []string{"der"}, nil
			}
			if strings.Contains(posTag, "SUB:") && strings.Contains(posTag, ":PLU:") {
				return []string{"Hunde"}, nil
			}
			if strings.Contains(posTag, "SUB:") {
				return []string{"Hund"}, nil
			}
			return []string{token.GetToken()}, nil
		},
	}
	det := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("der", strp("ART:DEF:NOM:SIN:MAS"), strp("der")), 0)
	noun := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Hund", strp("SUB:NOM:SIN:MAS"), strp("Hund")), 4)
	s := NewAgreementSuggestor2(synth, det, noun)
	all := s.GetSuggestionsFiltered(false)
	filtered := s.GetSuggestionsFiltered(true)
	require.NotEmpty(t, all)
	require.NotEmpty(t, filtered)
	// filtered must be a prefix tier: no more suggestions than unfiltered
	require.LessOrEqual(t, len(filtered), len(all))
	// at least one single-edit suggestion like "den Hund"
	require.Contains(t, filtered, "den Hund")
}

func TestAgreementSuggestor2_ProPosTemplate(t *testing.T) {
	// PRO:POS path uses proPosTemplates, not ART.
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil {
				return nil, nil
			}
			if strings.Contains(posTag, "PRO:POS:") && strings.Contains(posTag, ":AKK:") {
				return []string{"meinen"}, nil
			}
			if strings.Contains(posTag, "PRO:POS:") {
				return []string{"mein"}, nil
			}
			if strings.Contains(posTag, "SUB:") {
				return []string{"Hund"}, nil
			}
			return nil, nil
		},
	}
	det := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("mein", strp("PRO:POS:NOM:SIN:MAS:BEG"), strp("mein")), 0)
	noun := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Hund", strp("SUB:NOM:SIN:MAS"), strp("Hund")), 5)
	sugs := NewAgreementSuggestor2(synth, det, noun).GetSuggestions()
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs, "meinen Hund")
}

func TestReplaceSuggestorVars(t *testing.T) {
	got := replaceSuggestorVars(detTemplate, "SIN", "NEU", "AKK")
	// IND/DEF still present until getDetOrPronounSynth replaces it
	require.Contains(t, got, "SIN")
	require.Contains(t, got, "NEU")
	require.Contains(t, got, "AKK")
	require.NotContains(t, got, "PLU")
	require.NotContains(t, got, "MAS/FEM")
}

func strp(s string) *string { return &s }
