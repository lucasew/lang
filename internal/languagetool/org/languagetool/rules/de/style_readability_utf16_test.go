package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStyleRepeated_UTF16LengthGates(t *testing.T) {
	r := NewGermanStyleRepeatedWordRule(nil)
	r.TestCompoundWords = true
	// length < 3 UTF-16 → not part of word
	require.False(t, r.isPartOfWord("ab", "cd"))
	// same lengths with short words
	require.False(t, r.isPartOfWord("ab", "abc")) // tokenText len 3, test 2 — order swaps
}

func TestStyleRepeated_SecondPartSubstringUTF16(t *testing.T) {
	r := NewGermanStyleRepeatedWordRule(nil)
	r.IsCorrectSpell = func(word string) bool {
		return word == "haus" || word == "bau"
	}
	// "Spielhaus" starts with lower "spielhaus" prefix "spiel"? lowerTokenText of "Haus" = "haus"
	// isSecondPart: lower(test).startsWith("haus")? spielhaus.startsWith(haus)=false
	// endsWith haus? Spielhaus ends with haus? "Spielhaus" ends with "haus" yes
	// word = Spiel prefix via drop len("Haus")=4 → "Spiel" — not spelled; with s strip last unit
	// force correct: use test "Autobahn" token "bahn"
	r.IsCorrectSpell = func(word string) bool { return word == "Auto" || word == "auto" }
	require.True(t, r.isSecondPartOfWord("Autobahn", "bahn"))
}

func TestReadability_SyllablesUTF16(t *testing.T) {
	// German vowels including umlauts as single BMP units
	require.Equal(t, 1, simpleSyllablesCountDE("a"))
	require.Equal(t, 2, simpleSyllablesCountDE("Auto")) // A-u-to → A vowel, u with A is double?, t, o
	// Auto: A vowel n=1; u with A → lastDouble; t; o vowel lastDouble → n=2; o alone? lastDouble true so o increments → 2
	// Recheck: i=0 A vowel; i=1 u, cl=A → lastDouble=true no count; i=2 t; i=3 o vowel lastDouble → n=2
	require.Equal(t, 2, simpleSyllablesCountDE("Auto"))
	require.Equal(t, 1, simpleSyllablesCountDE("x")) // no vowel → 1
	require.Equal(t, 0, simpleSyllablesCountDE(""))
	// ü is vowel
	require.Equal(t, 1, simpleSyllablesCountDE("ü"))
	require.Equal(t, 2, simpleSyllablesCountDE("über")) // ü, e with ü not double, e counts → wait
	// ü vowel; b; e vowel (cl=b) count; r → 2
}

func TestIsFirstUpper_UTF16(t *testing.T) {
	require.True(t, isFirstUpper("Haus"))
	require.False(t, isFirstUpper("haus"))
	require.False(t, isFirstUpper(""))
	require.True(t, isFirstUpper("Über"))
}
