package fr

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// FrenchSynthesizer ports org.languagetool.synthesis.FrenchSynthesizer.
type FrenchSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// Java: burkinabè, koinè, épistémè allowed with final è.
var exceptionsEgrave = map[string]struct{}{
	"burkinabè": {},
	"koinè":     {},
	"épistémè":  {},
}

func NewFrenchSynthesizer(manual *synthesis.ManualSynthesizer) *FrenchSynthesizer {
	base := synthesis.NewBaseSynthesizer("fr", manual)
	// Java: super("fr/fr.sor", "/fr/french_synth.dict", "/fr/french_tags.txt", "fr")
	base.SorFileName = "fr/fr.sor"
	base.ResourceFileName = "/fr/french_synth.dict"
	base.TagFileName = "/fr/french_tags.txt"
	// Java virtual isException → Base removeExceptions on synthesize paths.
	base.IsExceptionFn = frenchIsException
	return &FrenchSynthesizer{BaseSynthesizer: base}
}

// frenchIsException is the Java FrenchSynthesizer.isException body (for Base.IsExceptionFn).
func frenchIsException(w string) bool {
	if strings.HasPrefix(w, "qq") {
		return true
	}
	if strings.HasSuffix(w, "è") {
		if _, ok := exceptionsEgrave[strings.ToLower(w)]; !ok {
			return true
		}
	}
	return false
}

// IsException ports FrenchSynthesizer.isException (filter synthesised forms).
func (s *FrenchSynthesizer) IsException(w string) bool {
	return frenchIsException(w)
}

// Synthesize / SynthesizeRE use Base paths; IsExceptionFn filters via RemoveExceptions.
// Keep method overrides so *FrenchSynthesizer implements the same surface as before
// (embedding already promotes Base methods; explicit aliases document Java override).
func (s *FrenchSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *FrenchSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*FrenchSynthesizer)(nil)
