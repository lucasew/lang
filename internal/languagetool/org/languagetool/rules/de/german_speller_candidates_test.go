package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterForLanguage_CH_ss(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	r.LanguageVariant = "CH"
	out := r.FilterForLanguage([]string{"Straße", "Maß"})
	require.Equal(t, []string{"Strasse", "Mass"}, out)
}

func TestFilterForLanguage_DropsLeadingDashAndSingleLetterTokens(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	out := r.FilterForLanguage([]string{"-Gratifikationskrisen", "Mafiosi s", "ok", "foo s."})
	require.Equal(t, []string{"ok"}, out)
}

func TestGetFilteredSuggestions_NoTaggerKeeps(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	// without TagPOS, POS-dependent filters must not invent drops
	in := []string{"Release Prozess", "groß Denken", "normal"}
	require.Equal(t, in, r.GetFilteredSuggestions(in))
}

func TestGetFilteredSuggestions_WithTagger(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	r.TagPOS = func(word string) []string {
		switch word {
		case "Release":
			return []string{"SUB:NOM:SIN:NEU"}
		case "Prozess":
			return []string{"SUB:NOM:SIN:MAS"}
		case "groß":
			return []string{"ADJ:PRD:GRU:SOL"}
		case "Denken":
			return []string{"SUB:NOM:SIN:NEU:INF"}
		default:
			return nil
		}
	}
	out := r.GetFilteredSuggestions([]string{"Release Prozess", "groß Denken", "keep me"})
	require.Equal(t, []string{"keep me"}, out)
}

func TestPostFilterGetSuggestions(t *testing.T) {
	out := postFilterGetSuggestions("foo", []string{"foo", "bar", "x yyy", "baz-", "qux"})
	require.Equal(t, []string{"bar", "qux"}, out)
	// allow trailing dash when input ends with dash
	out2 := postFilterGetSuggestions("foo-", []string{"bar-"})
	require.Equal(t, []string{"bar-"}, out2)
}

func TestGetCandidates_TwoPartSpaceAndS(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	// mock tokenizer: "gleichgroß" -> gleich + groß
	r.CompoundTokenize = func(word string) []string {
		if word == "gleichgroß" {
			return []string{"gleich", "groß"}
		}
		return []string{word}
	}
	cands := r.GetCandidates("gleichgroß")
	require.Contains(t, cands, "gleich groß")
	// no-s suffix path: parts[0] does not end with s → also "gleichsgroß"
	require.Contains(t, cands, "gleichsgroß")
}

func TestSortSuggestionByQuality_BoostsSpaceAndCase(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	out := r.SortSuggestionByQuality("Haus", []string{"Maus", "haus", "vor allem", "Auto"})
	// case-only + space first
	require.Equal(t, "haus", out[0])
	require.Equal(t, "vor allem", out[1])
}
