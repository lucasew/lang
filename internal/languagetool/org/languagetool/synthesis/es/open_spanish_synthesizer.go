package es

import (
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenSpanishSynthesizerFromDir loads Java SpanishSynthesizer resources from a
// resource directory containing es-ES_synth.dict (+ tags/manual when present).
func OpenSpanishSynthesizerFromDir(resourceDir string) *SpanishSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("es", resourceDir)
	if base == nil {
		return nil
	}
	if base.ResourceFileName == "" {
		base.ResourceFileName = SpanishSynthDict
	}
	if base.TagFileName == "" {
		base.TagFileName = SpanishTagsFile
	}
	if base.SorFileName == "" {
		base.SorFileName = SpanishSorFile
	}
	return &SpanishSynthesizer{BaseSynthesizer: base}
}

// OpenSpanishSynthesizerFromDictPath loads from the directory of es-ES_synth.dict.
func OpenSpanishSynthesizerFromDictPath(dictPath string) *SpanishSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenSpanishSynthesizerFromDir(filepath.Dir(dictPath))
}
