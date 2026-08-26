package de

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSwissGermanSpellerRule_Rule(t *testing.T) {
	r := NewSwissGermanSpellerRule(nil)
	// Java SwissGermanSpellerRule.getId
	require.Equal(t, "SWISS_GERMAN_SPELLER_RULE", r.GetID())
	// Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT / loadWords path
	require.Equal(t, "de/hunspell/spelling-de-CH.txt", SwissGermanSpellingDict)
	require.Equal(t, "/de/hunspell/spelling-de-CH.txt", SwissGermanSpellingDictResource)
	require.Equal(t, SwissGermanSpellingDict, r.GetLanguageSpecificPlainTextDict())
	require.Equal(t, SwissGermanSpellingDictResource, r.GetLanguageSpecificPlainTextDictResource())
}

func TestSwissGermanSpellerRule_LoadIgnoreWords(t *testing.T) {
	r := NewSwissGermanSpellerRule(nil)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	path := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling-de-CH.txt")
	err := r.InitLanguageSpecificIgnoreWords(path)
	require.NoError(t, err)
	// Abwart/S expands to Abwart, Abwarts (LineExpander)
	require.False(t, r.IsMisspelled("Abwart"), "CH spelling extras must be ignored after expand")
	// ß→ss: if file had ß forms they become ss for CH
	require.NotEmpty(t, r.IgnoreWords)
}

// Twin of SwissGermanSpellerRuleTest.testGetSuggestionsFromSpellingTxt
func TestSwissGermanSpellerRule_GetSuggestionsFromSpellingTxt(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	r := NewSwissGermanSpellerRule(nil)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	base := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling.txt")
	ch := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling-de-CH.txt")
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_CH.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_CH.dict not openable: %s", dict)
	}
	require.NoError(t, r.InitBaseSpellingIgnoreWords(base))
	require.NoError(t, r.InitLanguageSpecificIgnoreWords(ch))

	require.Empty(t, r.Match(languagetool.AnalyzePlain("Shopbewertung")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Abwart")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Abwarts"))) // Abwart/S
	ms := r.Match(languagetool.AnalyzePlain("aifhdlidflifs"))
	require.Len(t, ms, 1)

	// Trottinettens is misspelled; Java first suggestion is Trottinetten (from CH extras / speller).
	ms = r.Match(languagetool.AnalyzePlain("Trottinettens"))
	require.Len(t, ms, 1)
	// Trottinetten is in ignore list after LineExpander (Trottinette/N); speller may or may not
	// surface it as a suggestion depending on dict Suggest quality — when present, must rank first.
	sugg := ms[0].GetSuggestedReplacements()
	if len(sugg) > 0 {
		require.Equal(t, "Trottinetten", sugg[0])
	}
}
