package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPrepositionAndDeterminer(t *testing.T) {
	// consonant-start masculine: el
	require.Equal(t, "el ", GetPrepositionAndDeterminer("cotxe", "MS", ""))
	// vowel-start masculine: l'
	require.Equal(t, "l'", GetPrepositionAndDeterminer("amic", "MS", ""))
	// feminine vowel: l'
	require.Equal(t, "l'", GetPrepositionAndDeterminer("aigua", "FS", ""))
	// feminine consonant: la
	require.Equal(t, "la ", GetPrepositionAndDeterminer("casa", "FS", ""))
	// de + masc vowel
	require.Equal(t, "de l'", GetPrepositionAndDeterminer("amic", "MS", "de"))
	// a + fem consonant
	require.Equal(t, "a la ", GetPrepositionAndDeterminer("casa", "FS", "a"))
	// plural
	require.Equal(t, "els ", GetPrepositionAndDeterminer("cotxes", "MP", ""))
	// "ui" diphthong exception: home? h?ui[aeio]...
	// "hiena" - pMascNo matches hui? no. "hiena" starts with hi + e → pMascNo
	require.Equal(t, "el ", GetPrepositionAndDeterminer("hiena", "MS", ""))
}
