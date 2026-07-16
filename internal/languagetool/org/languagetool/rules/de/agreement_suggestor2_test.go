package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestAgreementSuggestor2(t *testing.T) {
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			// return lemma+tag suffix for visibility
			return []string{token.GetToken()}, nil
		},
	}
	det := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("der", strp("ART:DEF:NOM:SIN:MAS"), strp("der")), 0)
	noun := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Hund", strp("SUB:NOM:SIN:MAS"), strp("Hund")), 4)
	s := NewAgreementSuggestor2(synth, det, noun)
	got := s.GetSuggestions()
	require.NotEmpty(t, got)
	require.Contains(t, got, "der Hund")
}

func strp(s string) *string { return &s }
