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

// Twin: NERService.parseBuffer uses buffer.trim().split(" ") not Fields.
func TestParseBuffer_JavaTrimSpaceSplit(t *testing.T) {
	// double space → empty mid field skipped; both PERSON spans kept
	spans := ParseBuffer("  John/PERSON/0/4  Jane/PERSON/5/9  ")
	require.Len(t, spans, 2)
	// tab is not a delimiter for split(" ") — one token; parse from right still finds last PERSON
	spans = ParseBuffer("John/PERSON/0/4\tJane/PERSON/5/9")
	require.Len(t, spans, 1)
	require.Equal(t, 5, spans[0].GetStart())
	// NBSP not trimmed by String.trim — leading NBSP leaves token unparsable as PERSON start
	// (value still has NBSP prefix; slash parse may still find PERSON if structure intact)
	spans = ParseBuffer("\u00a0X/PERSON/0/1")
	require.Len(t, spans, 1) // trim leaves NBSP; token "\u00a0X/PERSON/0/1" still parses PERSON
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
