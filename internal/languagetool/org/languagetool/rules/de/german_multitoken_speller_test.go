package de

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanMultitokenSpeller_IsException(t *testing.T) {
	s := NewGermanMultitokenSpeller()
	require.True(t, s.IsException("Autos", "Auto"))
	require.True(t, s.IsException("foo-", "foo"))
	require.False(t, s.IsException("Haus", "Häuser"))
	require.False(t, s.IsException("Auto", "Autos"))
	// Java UTF-16 length-1: multi-byte prefix + trailing s (byte-slice would break)
	require.True(t, s.IsException("Müßigs", "Müßig"))
	// trailing non-s/- is not exception
	require.False(t, s.IsException("Müßige", "Müßig"))
}

func TestGermanMultitokenSpeller_IsExceptionStopsSuggestions(t *testing.T) {
	s := NewGermanMultitokenSpeller()
	// load a multiword phrase
	require.NoError(t, s.LoadWords(strings.NewReader("New York City\n")))
	// exact match → no suggestions
	require.Empty(t, s.GetSuggestions("New York City"))
	// exception: candidate without trailing s
	// put "Auto" multiword? need multi-token: "Foo Bar"
	require.NoError(t, s.LoadWords(strings.NewReader("Foo Bar\n")))
	// If original is "Foo Bars" and dict has "Foo Bar", isException may stop search
	// when candidates include "Foo Bar" and original ends with s matching candidate+s
	// stopSearching iterates candidates; IsException("Foo Bars", "Foo Bar") is true
	require.Empty(t, s.GetSuggestions("Foo Bars"))
}

func TestDiscoverAndLoadGermanMultitokenSpeller_NoPanic(t *testing.T) {
	s := DiscoverAndLoadGermanMultitokenSpeller()
	require.NotNil(t, s)
	require.NotNil(t, s.MultitokenSpeller)
}

func TestGermanMultitokenSpeller_ResourcePaths(t *testing.T) {
	// Java constructor Arrays.asList order
	require.Equal(t, []string{
		"/de/multitoken-suggest.txt",
		"/spelling_global.txt",
		"de/hunspell/spelling.txt",
	}, GermanMultitokenResourcePaths)
	require.NotNil(t, GermanMultitokenSpellerInstance)
}
