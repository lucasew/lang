package pt

import (
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenPortugueseSynthesizerFromDir loads Java PortugueseSynthesizer resources.
func OpenPortugueseSynthesizerFromDir(resourceDir string) *PortugueseSynthesizer {
	base := synthesis.OpenBaseSynthesizerFromDir("pt", resourceDir)
	if base == nil {
		return nil
	}
	if base.ResourceFileName == "" {
		base.ResourceFileName = PortugueseSynthDict
	}
	if base.TagFileName == "" {
		base.TagFileName = PortugueseTagsFile
	}
	if base.SorFileName == "" {
		base.SorFileName = PortugueseSorFile
	}
	return &PortugueseSynthesizer{BaseSynthesizer: base}
}

// OpenPortugueseSynthesizerFromDictPath loads from portuguese_synth.dict directory.
func OpenPortugueseSynthesizerFromDictPath(dictPath string) *PortugueseSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenPortugueseSynthesizerFromDir(filepath.Dir(dictPath))
}
