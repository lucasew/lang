package de

import (
	"path/filepath"
	"runtime"
	"testing"

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
