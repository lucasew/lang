package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAgreementCategoryString(t *testing.T) {
	s := AgreementCategoryString("NOM", "SIN", "MAS", "DEF", nil)
	require.Equal(t, "NOM/SIN/MAS/DEF", s)
	s2 := AgreementCategoryString("NOM", "SIN", "MAS", "DEF", map[GrammarCategory]bool{CatGenus: true})
	require.Equal(t, "NOM/SIN/DEF", s2)
}

// Java AgreementTools.getAgreementSOLCategories — only :SOL readings.
func TestGetAgreementSOLCategories(t *testing.T) {
	pSol := "ADJ:NOM:SIN:NEU:GRU:SOL"
	pInd := "ADJ:NOM:SIN:NEU:GRU:IND"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("rotes", &pSol, strPtr("rot")), 0)
	// add non-SOL reading that must be ignored
	tok.AddReading(languagetool.NewAnalyzedToken("rotes", &pInd, strPtr("rot")), "")
	cats := GetAgreementSOLCategories(tok, nil)
	require.NotEmpty(t, cats, "SOL reading must yield categories")
	// non-SOL-only token
	tok2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("rote", &pInd, strPtr("rot")), 0)
	require.Empty(t, GetAgreementSOLCategories(tok2, nil))
}
