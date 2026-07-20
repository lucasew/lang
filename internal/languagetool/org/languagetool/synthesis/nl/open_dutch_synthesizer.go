package nl

import (
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenDutchSynthesizerFromDir loads Java DutchSynthesizer resources.
func OpenDutchSynthesizerFromDir(resourceDir string) *DutchSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("nl", resourceDir)
	if base == nil {
		return nil
	}
	if base.ResourceFileName == "" {
		base.ResourceFileName = "/nl/dutch_synth.dict"
	}
	if base.TagFileName == "" {
		base.TagFileName = "/nl/dutch_tags.txt"
	}
	if base.SorFileName == "" {
		base.SorFileName = "/nl/nl.sor"
	}
	return &DutchSynthesizer{BaseSynthesizer: base}
}

// OpenDutchSynthesizerFromDictPath loads from dutch_synth.dict directory.
func OpenDutchSynthesizerFromDictPath(dictPath string) *DutchSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenDutchSynthesizerFromDir(filepath.Dir(dictPath))
}
