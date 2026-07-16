package dumpcheck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlainTextSentenceSource(t *testing.T) {
	in := "# source: http://example.com\nThis is a long enough sentence here.\nshort\nAnother reasonably long sentence line.\n"
	src := NewPlainTextSentenceSource(strings.NewReader(in))
	require.True(t, src.HasNext())
	s, err := src.Next()
	require.NoError(t, err)
	require.Equal(t, "This is a long enough sentence here.", s.GetText())
	require.Equal(t, "http://example.com", s.GetSource())
	s, err = src.Next()
	require.NoError(t, err)
	require.Equal(t, "Another reasonably long sentence line.", s.GetText())
	require.False(t, src.HasNext())
}
