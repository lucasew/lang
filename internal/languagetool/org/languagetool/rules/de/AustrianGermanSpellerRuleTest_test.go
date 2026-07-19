package de

import (
	"path/filepath"
	"runtime"
	"testing"

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
