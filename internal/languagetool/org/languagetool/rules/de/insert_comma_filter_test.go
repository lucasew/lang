package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestInsertCommaFilter_TwoToken(t *testing.T) {
	f := NewInsertCommaFilter()
	got := f.Suggest("hoffe es")
	require.Equal(t, []string{"hoffe, es"}, got)
}

func TestInsertCommaFilter_ThreeTokenWithoutTagger(t *testing.T) {
	// fail-closed: no POS invent for 3-token path
	f := NewInsertCommaFilter()
	require.Empty(t, f.Suggest("hoffe es geht"))
}

func TestInsertCommaFilter_ThreeTokenWithTagger(t *testing.T) {
	f := NewInsertCommaFilter()
	f.TagToken = func(w string) []string {
		switch w {
		case "hoffe":
			return []string{"VER:1:SIN:PRÄ:NON"}
		case "es":
			return []string{"PRO:PER:NOM:SIN:3:NEU"}
		case "geht":
			return []string{"VER:3:SIN:PRÄ:SFT"}
		}
		return nil
	}
	got := f.Suggest("hoffe es geht")
	require.Equal(t, []string{"hoffe, es geht"}, got)
}

func TestInsertCommaFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.de.InsertCommaFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.de.InsertCommaFilter")
	m := rules.NewRuleMatch(rules.NewFakeRule("I"), nil, 0, 8, "msg")
	m.SetSuggestedReplacements([]string{"sag mal"})
	out := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"sag, mal"}, out.GetSuggestedReplacements())
}

func TestInsertCommaFilter_FourTokenDenkeDemonstrative(t *testing.T) {
	// Java: DENKE_ETC + PRO:DEM: + SUB: → "schätze, diese Krawatte passt"
	f := NewInsertCommaFilter()
	f.TagToken = func(w string) []string {
		switch w {
		case "schätze":
			return []string{"VER:1:SIN:PRÄ:NON"}
		case "diese":
			return []string{"PRO:DEM:NOM:SIN:FEM"}
		case "Krawatte":
			return []string{"SUB:NOM:SIN:FEM"}
		case "passt":
			return []string{"VER:3:SIN:PRÄ:SFT"}
		}
		return nil
	}
	require.Equal(t, []string{"schätze, diese Krawatte passt"}, f.Suggest("schätze diese Krawatte passt"))
}

func TestInsertCommaFilter_FourTokenVerProPerAdvInr(t *testing.T) {
	// Java: VER + PRO:PER + ADV:INR → "Weißt du, warum diese"
	f := NewInsertCommaFilter()
	f.TagToken = func(w string) []string {
		switch w {
		case "Weißt":
			return []string{"VER:2:SIN:PRÄ:NON"}
		case "du":
			return []string{"PRO:PER:NOM:SIN:2"}
		case "warum":
			return []string{"ADV:INR"}
		case "diese":
			return []string{"PRO:DEM:NOM:SIN:FEM"}
		}
		return nil
	}
	require.Equal(t, []string{"Weißt du, warum diese"}, f.Suggest("Weißt du warum diese"))
}

func TestInsertCommaFilter_FourTokenWithoutTagger(t *testing.T) {
	f := NewInsertCommaFilter()
	require.Empty(t, f.Suggest("schätze diese Krawatte passt"))
}
