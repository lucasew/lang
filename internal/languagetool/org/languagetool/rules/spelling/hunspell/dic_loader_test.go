package hunspell

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDicWords(t *testing.T) {
	dic := "3\nhello/ABC\nworld\n# comment\nfoo/X\n"
	words, err := LoadDicWords(strings.NewReader(dic))
	require.NoError(t, err)
	require.Equal(t, []string{"hello", "world", "foo"}, words)
	dict, err := NewMapHunspellDictionaryFromDic(strings.NewReader(dic))
	require.NoError(t, err)
	require.True(t, dict.Spell("hello"))
	require.False(t, dict.Spell("missing"))
}
