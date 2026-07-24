package ca

import (
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenCatalanSynthesizerFromDir loads Java CatalanSynthesizer resources (ca-ES_synth.dict).
// langCode selects INSTANCE_CAT / VAL / BAL verb regional tags (default ca-ES).
func OpenCatalanSynthesizerFromDir(resourceDir, langCode string) *CatalanSynthesizer {
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
	if langCode == "" {
		langCode = "ca-ES"
	}
	// Normalize short "ca" → ca-ES (Java default Catalan).
	if strings.EqualFold(langCode, "ca") {
		langCode = "ca-ES"
	}
	return &CatalanSynthesizer{BaseSynthesizer: base, LanguageCode: langCode}
}

// OpenCatalanSynthesizerFromDictPath loads from the directory of ca-ES_synth.dict.
func OpenCatalanSynthesizerFromDictPath(dictPath, langCode string) *CatalanSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenCatalanSynthesizerFromDir(filepath.Dir(dictPath), langCode)
}
