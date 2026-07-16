package de

// Twin of GermanToolsTest.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanTools_IsVowel(t *testing.T) {
	require.True(t, IsVowel('a'))
	require.True(t, IsVowel('Y'))
	require.True(t, IsVowel('A'))
	require.True(t, IsVowel('ö'))
	require.False(t, IsVowel('b'))
}
