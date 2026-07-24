package de

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanCompoundTokenizer(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	tok.AddWord("auto")
	tok.AddWord("bahn")
	got := tok.Tokenize("autobahn")
	require.Equal(t, []string{"auto", "bahn"}, got)
	// unknown stays whole
	require.Equal(t, []string{"xyzabc"}, tok.Tokenize("xyzabc"))
}

func TestGermanCompoundTokenizer_AllSplits(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	// clear extras noise for deterministic small lexicon
	tok.Words = map[string]struct{}{
		"arbeit":      {},
		"amt":         {},
		"s":           {}, // not used: MinPartLen 3
		"platz":       {},
		"arbeitplatz": {}, // whole form also in dict
	}
	// "arbeitsplatz" with arbeit + platz only if "s" not needed as separate
	// arbeit + splatz won't work; need arbeit + platz with "s" glued
	tok.AddWord("arbeits")
	tok.AddWord("platz")
	// autobahn-style two-way
	tok.AddWord("auto")
	tok.AddWord("bahn")
	// also "autob" + "ahn" not in dict

	splits := tok.AllSplits("autobahn")
	require.NotEmpty(t, splits)
	require.Equal(t, []string{"auto", "bahn"}, splits[0])

	// multi-partition: ab+cd and a+b+c if all in dict — use letter compounds
	tok2 := NewGermanCompoundTokenizer(true)
	tok2.Words = map[string]struct{}{
		"foo":    {},
		"bar":    {},
		"baz":    {},
		"foobar": {}, // longer part also known
	}
	// foobarbaz: foo+bar+baz and foobar+baz
	all := tok2.AllSplits("foobarbaz")
	require.GreaterOrEqual(t, len(all), 2)
	var joined []string
	for _, p := range all {
		joined = append(joined, strings.Join(p, "+"))
	}
	require.Contains(t, joined, "foo+bar+baz")
	require.Contains(t, joined, "foobar+baz")
}

// Twin: Maskerade exception is commented out in Java — not an active exception.
func TestCompoundTokenizer_MaskeradeNotException(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	_, ok := tok.Exceptions["maskerade"]
	require.False(t, ok, "Maskerade is commented out in GermanCompoundTokenizer.java")
}
