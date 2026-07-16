package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/OldSpellingRuleTest.java
// Surface CSV + short-inflection port (no full German synthesizer / Aho-Corasick).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOldSpellingRule(t *testing.T) {
	rule := NewOldSpellingRule(nil)

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
	check("In Rußlands Weiten", "Russlands")
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

func TestOldSpellingRule_GermanAT(t *testing.T) {
	rule := NewOldSpellingRuleAT(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Geschoß"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Erdgeschoß"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Erdgeschoßes"))))
}
