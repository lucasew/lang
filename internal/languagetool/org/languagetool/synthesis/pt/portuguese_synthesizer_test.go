package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortugueseSynthesizer(t *testing.T) {
	s := NewPortugueseSynthesizer(nil)
	require.Equal(t, "pt", s.LangShortCode)
	require.Equal(t, PortugueseSynthDict, s.ResourceFileName)
	require.Equal(t, PortugueseSorFile, s.GetSorFileName())
	require.Equal(t, "pt", INSTANCE.LangShortCode)
}
