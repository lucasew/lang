package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgreementCategoryString(t *testing.T) {
	s := AgreementCategoryString("NOM", "SIN", "MAS", "DEF", nil)
	require.Equal(t, "NOM/SIN/MAS/DEF", s)
	s2 := AgreementCategoryString("NOM", "SIN", "MAS", "DEF", map[GrammarCategory]bool{CatGenus: true})
	require.Equal(t, "NOM/SIN/DEF", s2)
}
