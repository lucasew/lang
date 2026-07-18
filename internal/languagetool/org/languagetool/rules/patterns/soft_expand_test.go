package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSoftExpandBackrefs(t *testing.T) {
	require.Equal(t, "combined", softExpandBackrefs(`\1`, []string{"combined", "together"}))
	require.Equal(t, "combined together", softExpandBackrefs(`\1 \2`, []string{"combined", "together"}))
	require.Equal(t, "x", softExpandBackrefs(`\1`, []string{"x"}))
}

func TestExtractSuggestionsBackref(t *testing.T) {
	msg := `'\1 \2' is redundant. Use <suggestion>\1</suggestion>`
	clean, suggs := extractSuggestions(msg)
	require.Equal(t, []string{`\1`}, suggs, "clean=%q", clean)
}
