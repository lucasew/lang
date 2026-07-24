package fr

import (
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenFrenchSynthesizerFromDir loads Java FrenchSynthesizer resources (isException filter).
func OpenFrenchSynthesizerFromDir(resourceDir string) *FrenchSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("fr", resourceDir)
	if base == nil {
		return nil
	}
	if base.ResourceFileName == "" {
		base.ResourceFileName = "/fr/french_synth.dict"
	}
	if base.TagFileName == "" {
		base.TagFileName = "/fr/french_tags.txt"
	}
	if base.SorFileName == "" {
		base.SorFileName = "fr/fr.sor"
	}
	return &FrenchSynthesizer{BaseSynthesizer: base}
}

// OpenFrenchSynthesizerFromDictPath loads from the directory of french_synth.dict.
func OpenFrenchSynthesizerFromDictPath(dictPath string) *FrenchSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenFrenchSynthesizerFromDir(filepath.Dir(dictPath))
}
