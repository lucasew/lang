package de

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin: Maskerade exception is commented out in Java — not an active exception.
func TestCompoundTokenizer_MaskeradeNotException(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	_, ok := tok.Exceptions["maskerade"]
	require.False(t, ok, "Maskerade is commented out in GermanCompoundTokenizer.java")
}

// Twin: Sportshorts addException is commented out in Java.
func TestCompoundTokenizer_SportshortsNotException(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	_, ok := tok.Exceptions["sportshorts"]
	require.False(t, ok, "Sportshorts is commented out in GermanCompoundTokenizer.java")
}

// Helper: InputTooLongException path returns singleton (Java GermanCompoundTokenizer.tokenize catch).
func TestCompoundTokenizer_TooLongReturnsWhole(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	long := make([]byte, 71)
	for i := range long {
		long[i] = 'a'
	}
	got := tok.Tokenize(string(long))
	require.Equal(t, []string{string(long)}, got)
}

// Helper: splitter exposes InputTooLong for AllSplits (returns nil in wrapper).
func TestCompoundTokenizer_AllSplitsTooLong(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	long := make([]byte, 71)
	for i := range long {
		long[i] = 'a'
	}
	require.Nil(t, tok.AllSplits(string(long)))
}

// Helper: constructor exception casing preserved like Java asList.
func TestCompoundTokenizer_ExceptionCase(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	require.Equal(t, []string{"Alkohol", "abstinenz"}, tok.Tokenize("Alkoholabstinenz"))
	require.Equal(t, []string{"Flexo", "druck"}, tok.Tokenize("Flexodruck"))
}

// Sanity: embedded dict loaded (not empty invent). jWordSplitter 4.7 text dict
// (LanguageTool de.danielnaber.jwordsplitter.version) — ~58k lines / ~57k unique.
func TestCompoundTokenizer_EmbeddedDictLoaded(t *testing.T) {
	words := embeddedGermanDictionaryWords()
	require.GreaterOrEqual(t, len(words), 50000)
	_, ok := words["bahnhof"]
	require.True(t, ok)
	_, ok = words["nato"]
	require.True(t, ok)
	// 4.7 includes "schwarz" (needed for schwarzgrau strict split in GermanTagger)
	_, ok = words["schwarz"]
	require.True(t, ok)
}

// Sanity: errInputTooLong identity.
func TestInputTooLongError(t *testing.T) {
	sp := newGermanWordSplitter(false, map[string]struct{}{"aaa": {}})
	_, err := sp.splitWord("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") // 71
	require.True(t, errors.Is(err, errInputTooLong))
}

func TestCompoundTokenizer_MoreJavaOutcomes(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	cases := map[string][]string{
		"Bahnhofsuhr":      {"Bahnhofs", "uhr"},
		"natofreundlich":   {"nato", "freundlich"},
		"natofreundliches": {"nato", "freundliches"},
		"Firefox-Add-on":   {"Firefox", "", "Add-on"},
		"Alkoholabstinenz": {"Alkohol", "abstinenz"},
		"Flexodruck":       {"Flexo", "druck"},
		"Absolventen":      {"Absolventen"},
		"Haustür":          {"Haus", "tür"},
		"Autobahn":         {"Autobahn"},
		"Arbeitsplatz":     {"Arbeitsplatz"},
		"karamelligen":     {"karamelligen"},
		"Muttertag":        {"Muttertag"},
	}
	for in, want := range cases {
		require.Equal(t, want, tok.Tokenize(in), in)
	}
}
