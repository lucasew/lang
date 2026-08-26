package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java PortugueseSynthesizer.RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME.
const (
	PortugueseSynthDict = "/pt/portuguese_synth.dict"
	PortugueseTagsFile  = "/pt/portuguese_tags.txt"
	PortugueseSorFile   = "/pt/pt.sor"
)

// PortugueseSynthesizer ports org.languagetool.synthesis.pt.PortugueseSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type PortugueseSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// INSTANCE ports PortugueseSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *PortugueseSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *PortugueseSynthesizer) {
	INSTANCE = s
}

// NewPortugueseSynthesizer ports private PortugueseSynthesizer() / Language ctor.
// Java: super(SOR_FILE_NAME, RESOURCE_FILENAME, TAGS_FILE_NAME, "pt")
func NewPortugueseSynthesizer(manual *synthesis.ManualSynthesizer) *PortugueseSynthesizer {
	base := synthesis.NewBaseSynthesizer("pt", manual)
	base.SorFileName = PortugueseSorFile
	base.ResourceFileName = PortugueseSynthDict
	base.TagFileName = PortugueseTagsFile
	return &PortugueseSynthesizer{BaseSynthesizer: base}
}

// GetSorFileName returns Java SOR_FILE_NAME (/pt/pt.sor).
func (s *PortugueseSynthesizer) GetSorFileName() string {
	return PortugueseSorFile
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *PortugueseSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *PortugueseSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*PortugueseSynthesizer)(nil)
