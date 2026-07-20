package de

import "path/filepath"

// OpenGermanSynthesizerFromDictPath loads resources from the directory of german_synth.dict.
// Ports Language.createDefaultSynthesizer() → GermanSynthesizer.INSTANCE.
func OpenGermanSynthesizerFromDictPath(dictPath string) *GermanSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenGermanSynthesizerFromDir(filepath.Dir(dictPath))
}
