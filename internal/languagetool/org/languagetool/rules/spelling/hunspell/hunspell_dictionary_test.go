package hunspell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapHunspellDictionary(t *testing.T) {
	d := NewMapHunspellDictionary([]string{"hello", "world"})
	require.True(t, d.Spell("hello"))
	require.False(t, d.Spell("helo"))
	d.Add("helo")
	require.True(t, d.Spell("helo"))
	d.SetSuggestions("helo", []string{"hello"})
	require.Equal(t, []string{"hello"}, d.Suggest("helo"))
	require.NoError(t, d.Close())
	require.True(t, d.IsClosed())
	require.False(t, d.Spell("hello"))
}
