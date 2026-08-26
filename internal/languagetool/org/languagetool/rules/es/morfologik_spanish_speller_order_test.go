package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

func TestMorfologikSpanishSpellerRule_IgnoreTaggedWords(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	require.True(t, r.IgnoreTaggedWords)
	require.Equal(t, "MORFOLOGIK_RULE_ES", r.GetID())
	require.Equal(t, "/es/es-ES.dict", r.GetFileName())
}

func TestOrderSpanishSuggestions_RemoveAndPrefix(t *testing.T) {
	// remove slang truncations; drop "anti foo" style splits
	// Java PARTICULA_FINAL is the second token ("que"/"cual")
	got := orderSpanishSuggestions([]string{"casa", "compu", "anti virus", "foo s", "casa que"}, "caza")
	require.Contains(t, got, "casa")
	require.NotContains(t, got, "compu")
	require.NotContains(t, got, "anti virus")
	require.NotContains(t, got, "foo s")
	// particle final moved front
	require.Equal(t, "casa que", got[0])
}

func TestOrderSpanishSuggestions_DiacriticPrefer(t *testing.T) {
	got := orderSpanishSuggestions([]string{"mapa", "papá", "papa"}, "papa")
	// diacritic-equivalent forms ahead of unrelated "mapa"
	require.Equal(t, "mapa", got[len(got)-1])
	require.Contains(t, got[:2], "papá")
	require.Contains(t, got[:2], "papa")
	require.Equal(t, "papa", tools.RemoveDiacritics("papá"))
}

func TestMorfologikSpanishSpellerRule_MatchFiltersSuggestions(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(SpanishSpellerDict, 1)
	sp.AddWord("casa")
	sp.Suggestions["caza"] = []string{"compu", "casa", "anti algo"}
	r := NewMorfologikSpanishSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("caza")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	sugs := m[0].GetSuggestedReplacements()
	require.Contains(t, sugs, "casa")
	require.NotContains(t, sugs, "compu")
	require.NotContains(t, sugs, "anti algo")
}

func TestOrderSpanishSuggestions_PronounVerbWithTagPOS(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	r.TagPOS = func(word string) []string {
		if word == "come" {
			return []string{"VMIP3S0"}
		}
		return nil
	}
	got := r.orderSpanishSuggestions([]string{"casa", "me come", "mapa"}, "xxx")
	require.Equal(t, "me come", got[0])
	require.Contains(t, got, "casa")
}

func TestOrderSpanishSuggestions_PronounWithoutTaggerNoMove(t *testing.T) {
	// without TagPOS, pronoun splits are not reordered (fail-closed, no invent)
	got := orderSpanishSuggestions([]string{"casa", "me come"}, "xxx")
	require.Equal(t, "casa", got[0])
	require.Contains(t, got, "me come")
}

func TestAdditionalTopSpanish_CamelCase(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	r.IsMisspelled = func(w string) bool { return false }
	got := r.additionalTopSpanishSuggestions("GuardaChuva")
	require.Equal(t, []string{"Guarda Chuva"}, got)
}

func TestAdditionalTopSpanish_Digits(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	require.Empty(t, r.additionalTopSpanishSuggestions("casa2"))
	r.TagPOS = func(w string) []string {
		if w == "casa" {
			return []string{"NCMS000"}
		}
		return nil
	}
	require.Equal(t, []string{"casa 2"}, r.additionalTopSpanishSuggestions("casa2"))
}
