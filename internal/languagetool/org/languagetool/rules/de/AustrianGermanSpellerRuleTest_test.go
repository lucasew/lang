package de

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAustrianGermanSpellerRule_Rule(t *testing.T) {
	r := NewAustrianGermanSpellerRule(nil)
	// Java AustrianGermanSpellerRule.getId
	require.Equal(t, "AUSTRIAN_GERMAN_SPELLER_RULE", r.GetID())
	// Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT / loadWords path
	require.Equal(t, "de/hunspell/spelling-de-AT.txt", AustrianGermanSpellingDict)
	require.Equal(t, "/de/hunspell/spelling-de-AT.txt", AustrianGermanSpellingDictResource)
	require.Equal(t, AustrianGermanSpellingDict, r.GetLanguageSpecificPlainTextDict())
	require.Equal(t, AustrianGermanSpellingDictResource, r.GetLanguageSpecificPlainTextDictResource())
	// Soft base without dict
	require.False(t, r.IsMisspelled("Haus"))
}

func TestAustrianGermanSpellerRule_LoadIgnoreWords(t *testing.T) {
	r := NewAustrianGermanSpellerRule(nil)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	path := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling-de-AT.txt")
	err := r.InitLanguageSpecificIgnoreWords(path)
	require.NoError(t, err)
	// Word from AT extras file (first non-comment entry)
	require.False(t, r.IsMisspelled("abendäße"), "AT spelling extras must be ignored")
	// Without dict, unknown non-ignore still fail-open false
	require.False(t, r.IsMisspelled("xyzzyNotInFile"))
}

func TestLoadSpellingWordList_Comments(t *testing.T) {
	// smoke: load real AT file
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	path := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling-de-AT.txt")
	words, err := LoadSpellingWordList(path)
	require.NoError(t, err)
	require.NotEmpty(t, words)
	require.Contains(t, words, "abendäße")
}

// Twin of AustrianGermanSpellerRuleTest.testGetSuggestionsFromSpellingTxt
func TestAustrianGermanSpellerRule_GetSuggestionsFromSpellingTxt(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	r := NewAustrianGermanSpellerRule(nil)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Base + AT extras (Java SpellingCheckRule + AustrianGermanSpellerRule.init)
	base := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling.txt")
	at := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling-de-AT.txt")
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_AT.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_AT.dict not openable: %s", dict)
	}
	require.NoError(t, r.InitBaseSpellingIgnoreWords(base))
	require.NoError(t, r.InitLanguageSpecificIgnoreWords(at))

	// Shopbewertung / Wahlzuckerl from spelling lists → 0 matches
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Shopbewertung")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Wahlzuckerl")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Wahlzuckerls"))) // /NS expand
	// nonsense → misspelled
	ms := r.Match(languagetool.AnalyzePlain("aifhdlidflifs"))
	require.Len(t, ms, 1)
}
