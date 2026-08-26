package de

import (
	"os"
	"path/filepath"
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
		// Prefer full DE resource dir (added.txt / removed / do-not-synthesize + synth dict).
		// Fallback: third_party/german-pos-dict (Maven de.danielnaber:german-pos-dict, same as Java).
		for _, root := range germanSynthesizerResourceCandidates() {
			if root == "" {
				continue
			}
			if gs := synthde.OpenGermanSynthesizerFromDir(root); gs != nil {
				germanSynth = gs
				return
			}
			// No full German open — try base only (no German filters).
			if germanSynthBase == nil {
				germanSynthBase = synthesis.OpenBaseSynthesizerFromDir("de", root)
			}
		}
	})
	return germanSynth
}

// germanSynthesizerResourceCandidates ports Java classpath lookup for /de/german_synth.dict.
// Order matches other DE discovery: inspiration module resource, then third_party german-pos-dict.
func germanSynthesizerResourceCandidates() []string {
	var out []string
	if root := DiscoverGermanResourceDir(); root != "" {
		out = append(out, root)
	}
	// Walk-up third_party/german-pos-dict (Maven de.danielnaber:german-pos-dict; same as Java).
	const rel = "third_party/german-pos-dict/org/languagetool/resource/de"
	if p := walkUpRelDir(rel); p != "" {
		if len(out) == 0 || out[0] != p {
			out = append(out, p)
		}
	}
	return out
}

// walkUpRelDir finds rel by walking upward from CWD (same pattern as DiscoverGermanResourceDir).
func walkUpRelDir(rel string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
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
