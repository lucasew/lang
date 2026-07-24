package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceBundleWithFallback(t *testing.T) {
	r := NewResourceBundleWithFallback(
		MessageBundle{"a": "", "b": "from-main"},
		MessageBundle{"a": "from-fallback", "b": "fb", "c": "only-fb"},
	)
	require.Equal(t, "from-fallback", r.GetString("a"))
	require.Equal(t, "from-main", r.GetString("b"))
	require.Equal(t, "only-fb", r.GetString("c"))
	keys := r.GetKeys()
	require.ElementsMatch(t, []string{"a", "b"}, keys)
	r2 := NewResourceBundleWithFallback(
		MessageBundle{"a": "", "b": "from-main"},
		MessageBundle{"a": "from-fallback", "b": "fb", "c": "only-fb"},
	)
	require.True(t, r.Equal(r2))
	require.Equal(t, r.Hash(), r2.Hash())
}
