package synthesis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSorosSimple(t *testing.T) {
	// trivial rewrite: exact match
	s := NewSoros(`"0" zero; "1" one`, "en")
	require.Equal(t, "zero", s.Run("0"))
	require.Equal(t, "one", s.Run("1"))
	require.Equal(t, "", s.Run("2"))
}

func TestSorosZeroStrip(t *testing.T) {
	// __numbertext__ adds zero stripping: 001 -> 1 then may fail without further rules
	s := NewSoros(`"1" one`, "en")
	// after zero strip 001 -> 1 -> one
	require.Equal(t, "one", s.Run("001"))
}
