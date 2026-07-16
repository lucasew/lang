package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRange_AndFragment(t *testing.T) {
	r := NewRange(0, 5, "en")
	require.True(t, r.Equal(NewRange(0, 5, "en")))
	require.False(t, r.Equal(NewRange(0, 5, "de")))
	f := NewFragmentWithLanguage("fr", "bonjour")
	require.Equal(t, "| fr: bonjour |", f.String())
}
