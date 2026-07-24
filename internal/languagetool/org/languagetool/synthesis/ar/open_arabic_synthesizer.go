package ar

import (
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenArabicSynthesizerFromDir loads Java ArabicSynthesizer resources from a
// resource directory containing arabic_synth.dict (+ tags/manual when present).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenArabicSynthesizerFromDir(resourceDir string) *ArabicSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("ar", resourceDir)
	if base == nil {
		return nil
	}
	if base.ResourceFileName == "" {
		base.ResourceFileName = ArabicSynthDict
	}
	if base.TagFileName == "" {
		base.TagFileName = ArabicTagsFile
	}
	return &ArabicSynthesizer{
		BaseSynthesizer: base,
		tagmanager:      nil, // lazy via tm()
	}
}

// OpenArabicSynthesizerFromDictPath loads resources from the directory of arabic_synth.dict.
// Ports Language.createDefaultSynthesizer() → ArabicSynthesizer (getPosTagCorrection).
func OpenArabicSynthesizerFromDictPath(dictPath string) *ArabicSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenArabicSynthesizerFromDir(filepath.Dir(dictPath))
}
