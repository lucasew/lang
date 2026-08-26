package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/OldSpellingRuleTest.java
//
// Java:
//   Language germanDE = Languages.getLanguageForShortCode("de-DE");
//   OldSpellingRule rule = new OldSpellingRule(JLanguageTool.getMessageBundle(), germanDE);
//   JLanguageTool lt = new JLanguageTool(germanDE);
//   rule.match(lt.getAnalyzedSentence(...))
//
// SpellingData expands ß→ss inflections via GermanSynthesizer.INSTANCE.synthesizeForPosTags
// (requires official german_synth.dict from german-pos-dict, same as Java classpath).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// oldSpellingDE ports static OldSpellingRuleTest setup for de-DE.
func oldSpellingDE(t *testing.T) (*OldSpellingRule, *languagetool.JLanguageTool) {
	t.Helper()
	// Ensure SpellingData expand path has GermanSynthesizer (Java always does).
	require.NotNil(t, openDiscoveredGermanSynthesizer(),
		"german_synth.dict required for SpellingData ß→ss expansion (Java GermanSynthesizer.INSTANCE); install german-pos-dict into resource/de")
	m, _ := loadOldSpelling()
	require.Equal(t, "Russlands", m["Rußlands"],
		"SpellingData must expand Rußland→Rußlands via synthesizer like Java")
	// Java: Languages.getLanguageForShortCode("de-DE"); new JLanguageTool(germanDE)
	lt := languagetool.NewJLanguageTool("de-DE")
	// Java: new OldSpellingRule(JLanguageTool.getMessageBundle(), germanDE)
	rule := NewOldSpellingRule(nil)
	return rule, lt
}

// TestOldSpellingRule_Test ports OldSpellingRuleTest.test().
func TestOldSpellingRule_Test(t *testing.T) {
	rule, lt := oldSpellingDE(t)

	check := func(sentence string, suggs ...string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", sentence)
		matches := rule.Match(sent)
		require.Equal(t, 1, len(matches), "sentence %q matches=%v", sentence, matches)
		require.Equal(t, suggs, matches[0].GetSuggestedReplacements(), "sentence %q", sentence)
	}
	no := func(sentence string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", sentence)
		matches := rule.Match(sent)
		require.Equal(t, 0, len(matches), "unexpected match for %q: %v", sentence, matches)
	}
	assertMatch := func(sentence string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", sentence)
		require.Equal(t, 1, len(rule.Match(sent)), "sentence %q", sentence)
	}

	// Java OldSpellingRuleTest.test matrix (suggestions via getSuggestedReplacements().toString())
	check("Ein Kuß", "Kuss")
	check("Das Corpus delicti", "Corpus Delicti")
	// Inflected ß forms via SpellingData + GermanSynthesizer (always 1 match in Java)
	check("In Rußlands Weiten", "Russlands")
	check("Hot pants", "Hotpants")
	check("Ich muß los", "muss")
	check("schwarzweißmalen", "schwarzweiß malen", "schwarz-weiß malen")
	check("geschneuzt", "geschnäuzt")
	check("naß machen", "nassmachen")
	check("Midlife-crisis", "Midlife-Crisis", "Midlifecrisis")
	check("Schluß", "Schluss")
	no("schluß") // case-sensitive: lowercase not listed for sentence-start-only pairs
	check("Schloß", "Schloss")
	check("radfahren", "Rad fahren")
	check("Photo", "Foto")
	check("Geschoß", "Geschoss")
	check("Erdgeschoß", "Erdgeschoss")
	check("Erdgeschoßes", "Erdgeschosses")

	no("In Russland")
	no("In Russlands Weiten")
	no("Schlüsse")
	no("Schloß Holte")
	no("in Schloß Holte")
	no("Schloß Holte ist")
	no("Asse")
	no("Photons") // not "Photo" substring match
	no("Photon")
	no("Des Photons")
	no("Photons ")
	no("Hallo Herr Naß")
	no("Hallo Hr. Naß")
	no("Hallo Frau Naß")
	no("Hallo Fr. Naß")
	no("Fr. Naß")
	no("Dr. Naß")
	no("Prof. Naß")
	no("Bell Telephone")
	no("Telephone Company")
	assertMatch("Naß ist das Wasser")
	assertMatch("Läßt du das bitte")
	no("Das mögliche Bestehenbleiben")
	no("Das mögliche Bloßstrampeln verhindern.")
	assertMatch("Bloßstrampeln konnte er sich nicht.")
}

// TestOldSpellingRule_GermanAT ports OldSpellingRuleTest.testGermanAT.
func TestOldSpellingRule_GermanAT(t *testing.T) {
	// Java: Languages.getLanguageForShortCode("de-AT"); new OldSpellingRule(..., germanAT)
	rule := NewOldSpellingRuleAT(nil)
	lt := languagetool.NewJLanguageTool("de-AT")
	// Special case for Austria because of the pronunciation:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Geschoß"))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Erdgeschoß"))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Erdgeschoßes"))))
}

// TestOldSpellingRule_ExpandFormsIntegration: SpellingData ß→ss expand contract (Java L68–75).
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
	require.Equal(t, "Russlands", sd.Map()["Rußlands"])
	neu, ok := lookupOldSpelling("Rußlands", sd.Map())
	require.True(t, ok)
	require.Equal(t, "Russlands", neu)
}

// TestOldSpellingRule_ATStartPositionOccupies: addMatch may skip (AT) but still occupies hit.begin.
func TestOldSpellingRule_ATStartPositionOccupies(t *testing.T) {
	rule := NewOldSpellingRuleAT(nil)
	lt := languagetool.NewJLanguageTool("de-AT")
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Geschoß"))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Erdgeschoß"))))
}

// TestOldSpellingRule_IgnoreMatchTwins: ignoreMatch boundary/title/Telephone extras.
func TestOldSpellingRule_IgnoreMatchTwins(t *testing.T) {
	rule, lt := oldSpellingDE(t)
	no := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence(s))), s)
	}
	no("Photons")
	no("Des Photons")
	no("Hallo Herr Naß")
	no("Prof. Naß")
	no("Bell Telephone")
	no("Telephone Company")
	// Sentence-start Läßt
	ms := rule.Match(lt.GetAnalyzedSentence("Läßt du das bitte"))
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Lässt"}, ms[0].GetSuggestedReplacements())
	// Suggestion message path for ß→ss covered words
	ms2 := rule.Match(lt.GetAnalyzedSentence("Ein Kuß"))
	require.Equal(t, 1, len(ms2))
	require.Equal(t, "Kuss", ms2[0].GetSuggestedReplacements()[0])
}
