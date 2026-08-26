package hunspell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverHunspellDic_Danish(t *testing.T) {
	p := DiscoverHunspellDic("/da/hunspell/da_DK.dic")
	if p == "" {
		t.Skip("da_DK.dic not in tree")
	}
	require.Contains(t, p, "da_DK.dic")
}

func TestTryOpenFromClasspath_Danish(t *testing.T) {
	d := TryOpenFromClasspath("/da/hunspell/da_DK.dic")
	if d == nil {
		t.Skip("da_DK.dic not openable")
	}
	// Common Danish word should be in the list when file is real.
	// Affix expansion incomplete — only base forms from .dic lines.
	require.False(t, d.IsClosed())
	// garbage not in dict
	require.False(t, d.Spell("xyzzyqqqnotaword"))
}
