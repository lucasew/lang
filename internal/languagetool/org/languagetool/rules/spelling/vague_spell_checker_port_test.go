package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVagueSpellChecker(t *testing.T) {
	v := NewVagueSpellChecker()
	dict := map[string]bool{"hello": true, "world": true}
	v.Register("en", func(word string) bool { return dict[word] })
	require.True(t, v.IsValidWord("hello", "en"))
	require.True(t, v.IsValidWord("hello", "en-US"))
	require.False(t, v.IsValidWord("xyzzy", "en"))
	require.False(t, v.IsValidWord("hello", "de"))
}
