package ner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseBuffer(t *testing.T) {
	// token/TAG/from/to
	spans := ParseBuffer("John/PERSON/0/4 lives/O/5/10")
	require.Len(t, spans, 1)
	require.Equal(t, 0, spans[0].GetStart())
	require.Equal(t, 4, spans[0].GetEnd())

	require.Empty(t, ParseBuffer("foo/ORG/1/3"))
}

func TestRunNER(t *testing.T) {
	s := NewNERService("http://example.invalid/ner")
	s.Post = func(endpoint, formBody string) (string, error) {
		require.Contains(t, formBody, "input=")
		return "Ada/PERSON/0/3", nil
	}
	spans := s.RunNER("Ada")
	require.Len(t, spans, 1)
	require.Equal(t, "0-3", spans[0].String())
}

func TestNewSpanPanics(t *testing.T) {
	require.Panics(t, func() { NewSpan(5, 5) })
}
