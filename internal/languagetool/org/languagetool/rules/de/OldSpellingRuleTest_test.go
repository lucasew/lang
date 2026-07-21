package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/OldSpellingRuleTest.java
// Inflected ß→ss (e.g. Rußlands) when german_synth.dict is discoverable; else CSV-only fail-closed.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOldSpellingRule_Test(t *testing.T) {
	rule := NewOldSpellingRule(nil)
	m, _ := loadOldSpelling()

	check := func(sentence string, sugg string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, sugg, matches[0].GetSuggestedReplacements()[0], "sentence %q got %v", sentence, matches[0].GetSuggestedReplacements())
	}
	no := func(sentence string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(sentence))), "sentence %q", sentence)
	}

	check("Ein Kuß", "Kuss")
	check("Das Corpus delicti", "Corpus Delicti")
	// Base CSV form always present; genitive Rußlands only via synthesizer (Java SpellingData).
	check("In Rußland Weiten", "Russland")
	if _, ok := m["Rußlands"]; ok {
		check("In Rußlands Weiten", "Russlands")
	} else {
		no("In Rußlands Weiten") // fail-closed without invent when synth dict missing
	}
	check("Hot pants", "Hotpants")
	check("Ich muß los", "muss")
	matches := rule.Match(languagetool.AnalyzePlain("schwarzweißmalen"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"schwarzweiß malen", "schwarz-weiß malen"}, matches[0].GetSuggestedReplacements())

	check("Ich muß los", "muss") // duplicate ok
	// forms present in CSV
	check("Schluß", "Schluss")
	no("schluß") // case-sensitive: lowercase not listed
	check("Schloß", "Schloss")
	check("radfahren", "Rad fahren")
	check("Photo", "Foto")
	check("Geschoß", "Geschoss")
	check("Erdgeschoß", "Erdgeschoss")
	check("Erdgeschoßes", "Erdgeschosses")

	no("In Russland")
	no("In Russlands Weiten")
	no("Schloß Holte")
	no("in Schloß Holte")
	no("Photons")
	no("Photon")
	no("Hallo Herr Naß")
	no("Dr. Naß")
	no("Bell Telephone")

	check("Naß ist das Wasser", "Nass")
	check("Läßt du das bitte", "Lässt")
}

// End-to-end SpellingData expand path used by OldSpellingRule (mock synth, no invent).
func TestOldSpellingRule_ExpandFormsIntegration(t *testing.T) {
	content := "Rußland;Russland\n"
	expand := func(old string) []string {
		if old == "Rußland" {
			return []string{"Rußlands"}
		}
		return nil
	}
	sd, err := LoadSpellingDataBoth(content, "t.csv", expand)
	require.NoError(t, err)
	require.Equal(t, "Russlands", sd.Map["Rußlands"])
	// suggestion for expanded form is ß→ss on the form itself (Java)
	neu, ok := lookupOldSpelling("Rußlands", sd.Map)
	require.True(t, ok)
	require.Equal(t, "Russlands", neu)
}

func TestOldSpellingRule_GermanAT(t *testing.T) {
	rule := NewOldSpellingRuleAT(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Geschoß"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Erdgeschoß"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Erdgeschoßes"))))
}

// Twin Java ignoreMatch: substring boundary, titles, Prof.
func TestOldSpellingRule_IgnoreMatchTwins(t *testing.T) {
	rule := NewOldSpellingRule(nil)
	// Photons must not match Photo
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Photons"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Des Photons"))))
	// Title + Naß
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Hallo Herr Naß"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Prof. Naß"))))
	// Sentence-start Läßt
	ms := rule.Match(languagetool.AnalyzePlain("Läßt du das bitte"))
	require.Equal(t, 1, len(ms))
	require.Equal(t, "Lässt", ms[0].GetSuggestedReplacements()[0])
}
