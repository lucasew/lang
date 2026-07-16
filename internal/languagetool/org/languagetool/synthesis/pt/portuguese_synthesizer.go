package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"

const (
	PortugueseSynthDict = "/pt/portuguese_synth.dict"
	PortugueseTagsFile  = "/pt/portuguese_tags.txt"
	PortugueseSorFile   = "/pt/pt.sor"
)

// PortugueseSynthesizer ports org.languagetool.synthesis.pt.PortugueseSynthesizer.
type PortugueseSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// INSTANCE mirrors Java PortugueseSynthesizer.INSTANCE.
var INSTANCE = NewPortugueseSynthesizer(nil)

func NewPortugueseSynthesizer(manual *synthesis.ManualSynthesizer) *PortugueseSynthesizer {
	base := synthesis.NewBaseSynthesizer("pt", manual)
	base.ResourceFileName = PortugueseSynthDict
	base.TagFileName = PortugueseTagsFile
	return &PortugueseSynthesizer{BaseSynthesizer: base}
}

func (s *PortugueseSynthesizer) GetSorFileName() string { return PortugueseSorFile }
