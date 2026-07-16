package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecadeSpellingFilter(t *testing.T) {
	f := NewDecadeSpellingFilter()
	// 1990 → decade 90, century 19 → wiek XX
	msg := f.FormatMessage("lata {dekada}. wieku {wiek}", "1990")
	require.Equal(t, "lata 90. wieku XX", msg)
	// 2000 → century 20 → XXI
	msg = f.FormatMessage("{dekada}/{wiek}", "2000")
	require.Equal(t, "00/XXI", msg)
	require.Equal(t, "", f.FormatMessage("x", "19"))
}
