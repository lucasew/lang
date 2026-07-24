package pl

import (
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenPolishSynthesizerFromDir loads Java PolishSynthesizer resources from a
// resource directory containing polish_synth.dict (+ tags/manual when present).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenPolishSynthesizerFromDir(resourceDir string) *PolishSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("pl", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics.
	if base.ResourceFileName == "" {
		base.ResourceFileName = "/pl/polish_synth.dict"
	}
	if base.TagFileName == "" {
		base.TagFileName = "/pl/polish_tags.txt"
	}
	return &PolishSynthesizer{BaseSynthesizer: base}
}

// OpenPolishSynthesizerFromDictPath loads resources from the directory of polish_synth.dict.
// Ports Language.createDefaultSynthesizer() → PolishSynthesizer (with getPosTagCorrection).
func OpenPolishSynthesizerFromDictPath(dictPath string) *PolishSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenPolishSynthesizerFromDir(filepath.Dir(dictPath))
}
