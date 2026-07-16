package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMoreVariants(t *testing.T) {
	require.Equal(t, "Italian", Italian.Name)
	require.Equal(t, "Dutch", DutchNetherlands.GetName())
	require.Len(t, AllDutchVariants(), 2)
	require.Equal(t, "Polish", Polish.Name)
	require.True(t, ValencianCatalan.Valencian)
	require.Equal(t, "Russian", Russian.Name)
	require.True(t, IsNonSwissGerman("de-DE"))
	require.False(t, IsNonSwissGerman("de-CH"))
}
