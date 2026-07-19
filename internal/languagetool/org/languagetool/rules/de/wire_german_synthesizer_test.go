package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSynthesizeGermanRE_FailClosed(t *testing.T) {
	// Empty lemma always nil
	require.Nil(t, synthesizeGermanRE("", "VER:.*"))
}

func TestOpenDiscoveredGermanSynthesizer_NoPanic(t *testing.T) {
	// May be nil without german_synth.dict — must not panic
	_ = openDiscoveredGermanSynthesizer()
	_ = openDiscoveredGermanSynthBase()
}

func TestGermanSynthesizer_SynthesizeForPosTagsCase(t *testing.T) {
	// Unit-level: German SynthesizeForPosTags applies case filter (not invent).
	// Covered in synthesis/de; wire path only needs discovery smoke.
	require.NotPanics(t, func() {
		if gs := openDiscoveredGermanSynthesizer(); gs != nil {
			_ = gs.SynthesizeForPosTags("Haus", func(string) bool { return true })
		}
	})
}
