package ca

import (
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenCatalanSynthesizerFromDir loads Java CatalanSynthesizer resources (ca-ES_synth.dict).
func OpenCatalanSynthesizerFromDir(resourceDir string) *CatalanSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("ca", resourceDir)
	if base == nil {
		return nil
	}
	if base.ResourceFileName == "" {
		base.ResourceFileName = "/ca/ca-ES_synth.dict"
	}
	if base.TagFileName == "" {
		base.TagFileName = "/ca/ca-ES_tags.txt"
	}
	if base.SorFileName == "" {
		base.SorFileName = "/ca/ca.sor"
	}
	return &CatalanSynthesizer{BaseSynthesizer: base}
}

// OpenCatalanSynthesizerFromDictPath loads from the directory of ca-ES_synth.dict.
func OpenCatalanSynthesizerFromDictPath(dictPath string) *CatalanSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenCatalanSynthesizerFromDir(filepath.Dir(dictPath))
}
