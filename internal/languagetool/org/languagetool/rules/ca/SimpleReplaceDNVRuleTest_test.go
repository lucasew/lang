package ca

// Twin of SimpleReplaceDNVRuleTest — lemma path + synth for plurals (no surface plural invent).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceDNVRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceDNVRule(nil)
	rule.Synthesize = dnvTestSynth

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ella és molt incauta."))))

	matches := rule.Match(analyzeCALemma("L'arxipèleg.", map[string]languagetool.TokenTag{
		"arxipèleg": {POS: "NCMS000", Lemma: "arxipèleg"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "arxipèlag", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(analyzeCALemma("colmena", map[string]languagetool.TokenTag{
		"colmena": {POS: "NCFS000", Lemma: "colmena"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "buc", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "rusc", matches[0].GetSuggestedReplacements()[1])

	// plural surface, lemma colmena, POS plural → synth plurals
	matches = rule.Match(analyzeCALemma("colmenes", map[string]languagetool.TokenTag{
		"colmenes": {POS: "NCFP000", Lemma: "colmena"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "bucs", matches[0].GetSuggestedReplacements()[0])
	require.Contains(t, matches[0].GetSuggestedReplacements(), "ruscos")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "ruscs")

	matches = rule.Match(analyzeCALemma("afincaments", map[string]languagetool.TokenTag{
		"afincaments": {POS: "NCMP000", Lemma: "afincament"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "establiments", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "instal·lacions", matches[0].GetSuggestedReplacements()[1])

	matches = rule.Match(analyzeCALemma("Els arxipèlegs", map[string]languagetool.TokenTag{
		"arxipèlegs": {POS: "NCMP000", Lemma: "arxipèleg"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "arxipèlags", matches[0].GetSuggestedReplacements()[0])
}

// dnvTestSynth ports enough CatalanSynthesizer behavior for the twin test.
func dnvTestSynth(lemma, postag string) []string {
	// plural noun tags NC.P.*
	plural := strings.Contains(postag, "P") && !strings.Contains(postag, "SP") // crude
	// FreeLing NCMP000 / NCFP000
	if len(postag) >= 4 && (postag[2] == 'P' || (len(postag) > 3 && postag[3] == 'P')) {
		plural = true
	}
	if strings.HasPrefix(postag, "NC") && len(postag) >= 4 {
		// NCMS000 vs NCMP000 — gender at [2], number at [3]
		if postag[3] == 'P' {
			plural = true
		} else if postag[3] == 'S' {
			plural = false
		}
	}
	switch lemma {
	case "arxipèlag":
		if plural {
			return []string{"arxipèlags"}
		}
		return []string{"arxipèlag"}
	case "buc":
		if plural {
			return []string{"bucs"}
		}
		return []string{"buc"}
	case "rusc":
		if plural {
			return []string{"ruscos", "ruscs"}
		}
		return []string{"rusc"}
	case "establiment":
		if plural {
			return []string{"establiments"}
		}
		return []string{"establiment"}
	case "instal·lació":
		if plural {
			return []string{"instal·lacions"}
		}
		return []string{"instal·lació"}
	case "encebar":
		return []string{"encebéssiu"}
	default:
		return nil
	}
}
