package de

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	synthde "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/de"
)

// openDiscoveredGermanSynthesizer returns GermanSynthesizer.INSTANCE stand-in when
// german_synth.dict is discoverable (case filter + compound fallback + REMOVE).
// Nil when resources missing (fail-closed).
var (
	germanSynthOnce sync.Once
	germanSynth     *synthde.GermanSynthesizer
	// base fallback if OpenGermanSynthesizerFromDir fails but base dict opens
	// (should not happen if same dir; kept for tests that mock only base).
	germanSynthBase *synthesis.BaseSynthesizer
)

func openDiscoveredGermanSynthesizer() *synthde.GermanSynthesizer {
	germanSynthOnce.Do(func() {
		root := DiscoverGermanResourceDir()
		if root == "" {
			return
		}
		if gs := synthde.OpenGermanSynthesizerFromDir(root); gs != nil {
			germanSynth = gs
			return
		}
		// No full German open — try base only (no German filters).
		germanSynthBase = synthesis.OpenBaseSynthesizerFromDir("de", root)
	})
	return germanSynth
}

// openDiscoveredGermanSynthBase returns German synthesizer when present, else bare Base.
// Prefer openDiscoveredGermanSynthesizer for German-specific filters.
func openDiscoveredGermanSynthBase() *synthesis.BaseSynthesizer {
	if gs := openDiscoveredGermanSynthesizer(); gs != nil {
		return gs.BaseSynthesizer
	}
	germanSynthOnce.Do(func() {}) // ensure once ran
	return germanSynthBase
}

// synthesizeGermanRE ports language.getSynthesizer().synthesize(token, re, true)
// with GermanSynthesizer filters when resources are present.
func synthesizeGermanRE(lemma, postagRE string) []string {
	if lemma == "" {
		return nil
	}
	lem := lemma
	tok := languagetool.NewAnalyzedToken(lemma, nil, &lem)
	if gs := openDiscoveredGermanSynthesizer(); gs != nil {
		forms, err := gs.SynthesizeRE(tok, postagRE, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	if base := openDiscoveredGermanSynthBase(); base != nil {
		forms, err := base.SynthesizeRE(tok, postagRE, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	return nil
}
