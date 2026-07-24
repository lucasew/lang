package gl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// GalicianSRXSentenceTokenizer ports tokenizers.gl.GalicianSRXSentenceTokenizer.
type GalicianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewGalicianSRXSentenceTokenizer() *GalicianSRXSentenceTokenizer {
	return &GalicianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("gl"),
	}
}
