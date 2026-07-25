package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.tagging.de.AdjectiveTags static lists (Java asList order + content).

func TestAdjectiveTags_tagsForAdj(t *testing.T) {
	require.Equal(t, []string{"ADJ:PRD:GRU"}, tagsForAdj)
	require.Equal(t, 1, len(tagsForAdj))
}

func TestAdjectiveTags_tagsForAdjE(t *testing.T) {
	want := []string{
		"ADJ:AKK:PLU:FEM:GRU:SOL",
		"ADJ:AKK:PLU:MAS:GRU:SOL",
		"ADJ:AKK:PLU:NEU:GRU:SOL",
		"ADJ:AKK:SIN:FEM:GRU:DEF",
		"ADJ:AKK:SIN:FEM:GRU:IND",
		"ADJ:AKK:SIN:FEM:GRU:SOL",
		"ADJ:AKK:SIN:NEU:GRU:DEF",
		"ADJ:NOM:PLU:FEM:GRU:SOL",
		"ADJ:NOM:PLU:MAS:GRU:SOL",
		"ADJ:NOM:PLU:NEU:GRU:SOL",
		"ADJ:NOM:SIN:FEM:GRU:DEF",
		"ADJ:NOM:SIN:FEM:GRU:IND",
		"ADJ:NOM:SIN:FEM:GRU:SOL",
		"ADJ:NOM:SIN:MAS:GRU:DEF",
		"ADJ:NOM:SIN:NEU:GRU:DEF",
	}
	require.Equal(t, want, tagsForAdjE)
	require.Equal(t, 15, len(tagsForAdjE))
}

func TestAdjectiveTags_tagsForAdjEn(t *testing.T) {
	want := []string{
		"ADJ:AKK:PLU:FEM:GRU:DEF",
		"ADJ:AKK:PLU:FEM:GRU:IND",
		"ADJ:AKK:PLU:MAS:GRU:DEF",
		"ADJ:AKK:PLU:MAS:GRU:IND",
		"ADJ:AKK:PLU:NEU:GRU:DEF",
		"ADJ:AKK:PLU:NEU:GRU:IND",
		"ADJ:AKK:SIN:MAS:GRU:DEF",
		"ADJ:AKK:SIN:MAS:GRU:IND",
		"ADJ:AKK:SIN:MAS:GRU:SOL",
		"ADJ:DAT:PLU:FEM:GRU:DEF",
		"ADJ:DAT:PLU:FEM:GRU:IND",
		"ADJ:DAT:PLU:FEM:GRU:SOL",
		"ADJ:DAT:PLU:MAS:GRU:DEF",
		"ADJ:DAT:PLU:MAS:GRU:IND",
		"ADJ:DAT:PLU:MAS:GRU:SOL",
		"ADJ:DAT:PLU:NEU:GRU:DEF",
		"ADJ:DAT:PLU:NEU:GRU:IND",
		"ADJ:DAT:PLU:NEU:GRU:SOL",
		"ADJ:DAT:SIN:FEM:GRU:DEF",
		"ADJ:DAT:SIN:FEM:GRU:IND",
		"ADJ:DAT:SIN:MAS:GRU:DEF",
		"ADJ:DAT:SIN:MAS:GRU:IND",
		"ADJ:DAT:SIN:NEU:GRU:DEF",
		"ADJ:DAT:SIN:NEU:GRU:IND",
		"ADJ:GEN:PLU:FEM:GRU:DEF",
		"ADJ:GEN:PLU:FEM:GRU:IND",
		"ADJ:GEN:PLU:MAS:GRU:DEF",
		"ADJ:GEN:PLU:MAS:GRU:IND",
		"ADJ:GEN:PLU:NEU:GRU:DEF",
		"ADJ:GEN:PLU:NEU:GRU:IND",
		"ADJ:GEN:SIN:FEM:GRU:DEF",
		"ADJ:GEN:SIN:FEM:GRU:IND",
		"ADJ:GEN:SIN:MAS:GRU:DEF",
		"ADJ:GEN:SIN:MAS:GRU:IND",
		"ADJ:GEN:SIN:MAS:GRU:SOL",
		"ADJ:GEN:SIN:NEU:GRU:DEF",
		"ADJ:GEN:SIN:NEU:GRU:IND",
		"ADJ:GEN:SIN:NEU:GRU:SOL",
		"ADJ:NOM:PLU:FEM:GRU:DEF",
		"ADJ:NOM:PLU:FEM:GRU:IND",
		"ADJ:NOM:PLU:MAS:GRU:DEF",
		"ADJ:NOM:PLU:MAS:GRU:IND",
		"ADJ:NOM:PLU:NEU:GRU:DEF",
		"ADJ:NOM:PLU:NEU:GRU:IND",
	}
	require.Equal(t, want, tagsForAdjEn)
	require.Equal(t, 44, len(tagsForAdjEn))
}

func TestAdjectiveTags_tagsForAdjEr(t *testing.T) {
	want := []string{
		"ADJ:DAT:SIN:FEM:GRU:SOL",
		"ADJ:GEN:PLU:FEM:GRU:SOL",
		"ADJ:GEN:PLU:MAS:GRU:SOL",
		"ADJ:GEN:PLU:NEU:GRU:SOL",
		"ADJ:GEN:SIN:FEM:GRU:SOL",
		"ADJ:NOM:SIN:MAS:GRU:IND",
		"ADJ:NOM:SIN:MAS:GRU:SOL",
	}
	require.Equal(t, want, tagsForAdjEr)
	require.Equal(t, 7, len(tagsForAdjEr))
}

func TestAdjectiveTags_tagsForAdjEm(t *testing.T) {
	want := []string{
		"ADJ:DAT:SIN:MAS:GRU:SOL",
		"ADJ:DAT:SIN:NEU:GRU:SOL",
	}
	require.Equal(t, want, tagsForAdjEm)
	require.Equal(t, 2, len(tagsForAdjEm))
}

func TestAdjectiveTags_tagsForAdjEs(t *testing.T) {
	want := []string{
		"ADJ:AKK:SIN:NEU:GRU:IND",
		"ADJ:AKK:SIN:NEU:GRU:SOL",
		"ADJ:NOM:SIN:NEU:GRU:IND",
		"ADJ:NOM:SIN:NEU:GRU:SOL",
	}
	require.Equal(t, want, tagsForAdjEs)
	require.Equal(t, 4, len(tagsForAdjEs))
}
