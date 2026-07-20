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
	return &FrenchSynthesizer{BaseSynthesizer: base}
}

// IsException ports FrenchSynthesizer.isException (filter synthesised forms).
func (s *FrenchSynthesizer) IsException(w string) bool {
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

func (s *FrenchSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	forms, err := s.BaseSynthesizer.Synthesize(token, posTag)
	return s.filterExceptions(forms), err
}
func (s *FrenchSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	forms, err := s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
	return s.filterExceptions(forms), err
}

func (s *FrenchSynthesizer) filterExceptions(forms []string) []string {
	if len(forms) == 0 {
		return forms
	}
	var out []string
	for _, w := range forms {
		if !s.IsException(w) {
			out = append(out, w)
		}
	}
	return out
}

var _ synthesis.Synthesizer = (*FrenchSynthesizer)(nil)
