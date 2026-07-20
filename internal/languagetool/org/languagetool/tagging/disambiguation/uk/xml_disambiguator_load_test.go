package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaultUkrainianXmlDisambiguator(t *testing.T) {
	path := discoverUKDisambiguationXML()
	if path == "" {
		t.Skip("inspiration disambiguation.xml not found")
	}
	d := LoadDefaultUkrainianXmlDisambiguator()
	require.NotNil(t, d, "path=%s", path)

	h := NewUkrainianHybridDisambiguator()
	require.NotNil(t, h.Inner, "hybrid should wire XML disambiguator when resource present")
	require.NotNil(t, h.Chunker)
	require.NotNil(t, h.Simple)
}

func TestDiscoverUKDisambiguationXML(t *testing.T) {
	p := discoverUKDisambiguationXML()
	if p == "" {
		t.Skip("no inspiration tree")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "/uk/")
}
