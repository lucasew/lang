package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	detok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
)

// OpenGermanSynthesizerFromDir ports GermanSynthesizer.INSTANCE resource loading:
// german_synth.dict (+ tags/manual/sor) and GermanCompoundTokenizer.getStrictInstance().
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenGermanSynthesizerFromDir(resourceDir string) *GermanSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("de", resourceDir)
	if base == nil {
		return nil
	}
	gs := &GermanSynthesizer{BaseSynthesizer: base}
	// Java getCompoundForms: GermanCompoundTokenizer.getStrictInstance().tokenize
	// (embedded jWordSplitter dict + LT exceptions only — not hunspell inject).
	gs.StrictCompoundTokenize = strictCompoundTokenizeGetStrictInstance()
	return gs
}

// strictCompoundTokenizeGetStrictInstance ports GermanCompoundTokenizer.getStrictInstance().tokenize.
func strictCompoundTokenizeGetStrictInstance() func(lemma string) []string {
	return func(lemma string) []string {
		if lemma == "" {
			return nil
		}
		return detok.GetStrictInstance().Tokenize(lemma)
	}
}
