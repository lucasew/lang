package de

import (
	"os"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	detok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
)

// OpenGermanSynthesizerFromDir ports GermanSynthesizer.INSTANCE resource loading:
// german_synth.dict (+ tags/manual) and strict GermanCompoundTokenizer for getCompoundForms.
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenGermanSynthesizerFromDir(resourceDir string) *GermanSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("de", resourceDir)
	if base == nil {
		return nil
	}
	gs := &GermanSynthesizer{BaseSynthesizer: base}
	gs.StrictCompoundTokenize = strictCompoundTokenizeFromDir(resourceDir)
	return gs
}

// strictCompoundTokenizeFromDir builds GermanCompoundTokenizer.getStrictInstance-style split.
// Loads hunspell de_DE.dic when present so splits are lexicon-backed (no invent).
func strictCompoundTokenizeFromDir(resourceDir string) func(lemma string) []string {
	tok := detok.NewGermanCompoundTokenizer(true)
	if resourceDir != "" {
		// Prefer de_DE.dic; fall back to any de_*.dic under hunspell/
		candidates := []string{
			filepath.Join(resourceDir, "hunspell", "de_DE.dic"),
			filepath.Join(resourceDir, "hunspell", "de_AT.dic"),
			filepath.Join(resourceDir, "hunspell", "de_CH.dic"),
		}
		for _, p := range candidates {
			f, err := os.Open(p)
			if err != nil {
				continue
			}
			_ = tok.LoadHunspellDic(f)
			_ = f.Close()
			break
		}
	}
	return func(lemma string) []string {
		if lemma == "" {
			return nil
		}
		return tok.Tokenize(lemma)
	}
}
