package ga

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLeniteEclipse(t *testing.T) {
	require.Equal(t, "bhean", Lenite("bean"))
	require.Equal(t, "bhean", Lenite("bhean")) // already lenited
	require.True(t, stringsHasPrefix(Eclipse("bean"), "m"))
	require.True(t, IsVowel('á'))
	require.False(t, IsVowel('b'))
}

func TestFixSuffix(t *testing.T) {
	r := FixSuffix("dóireamhail")
	require.True(t, stringsHasSuffix(r.GetWord(), "iúil"))
	require.Contains(t, r.GetAppendTag(), "MorphError")
}

func TestDemutate(t *testing.T) {
	r := Demutate("bhean")
	require.Equal(t, "bean", r.GetWord())
}

func stringsHasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }
func stringsHasSuffix(s, p string) bool {
	return len(s) >= len(p) && s[len(s)-len(p):] == p
}
