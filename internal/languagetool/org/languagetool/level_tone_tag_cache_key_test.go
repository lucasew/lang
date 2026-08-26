package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLevelToneTagCacheKey(t *testing.T) {
	a := NewLevelToneTagCacheKey(LevelDefault, []ToneTag{ToneFormal, ToneClarity})
	b := NewLevelToneTagCacheKey(LevelDefault, []ToneTag{ToneClarity, ToneFormal}) // order independent
	require.True(t, a.Equal(b))
	require.Equal(t, a.String(), b.String())
	c := NewLevelToneTagCacheKey(LevelPicky, []ToneTag{ToneClarity, ToneFormal})
	require.False(t, a.Equal(c))
}
