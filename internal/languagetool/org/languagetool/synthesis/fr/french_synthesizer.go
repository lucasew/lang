package fr

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java FrenchSynthesizer resource paths (constructor super args).
const (
	FrenchSorFile      = "fr/fr.sor"
	FrenchSynthResource = "/fr/french_synth.dict"
	FrenchTagsFile     = "/fr/french_tags.txt"
)

// FrenchSynthesizer ports org.languagetool.synthesis.FrenchSynthesizer.
type FrenchSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// Java: private static final List<String> exceptionsEgrave = Arrays.asList("burkinabè", "koinè", "épistémè");
var exceptionsEgrave = map[string]struct{}{
	"burkinabè": {},
	"koinè":     {},
	"épistémè":  {},
}

// INSTANCE ports FrenchSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *FrenchSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *FrenchSynthesizer) {
	INSTANCE = s
}

// NewFrenchSynthesizer ports private FrenchSynthesizer() / deprecated Language ctor.
// Java: super("fr/fr.sor", "/fr/french_synth.dict", "/fr/french_tags.txt", "fr")
func NewFrenchSynthesizer(manual *synthesis.ManualSynthesizer) *FrenchSynthesizer {
	base := synthesis.NewBaseSynthesizer("fr", manual)
	base.SorFileName = FrenchSorFile
	base.ResourceFileName = FrenchSynthResource
	base.TagFileName = FrenchTagsFile
	// Java virtual isException → Base removeExceptions on synthesize paths.
	base.IsExceptionFn = frenchIsException
	return &FrenchSynthesizer{BaseSynthesizer: base}
}

// frenchIsException is the Java FrenchSynthesizer.isException body (for Base.IsExceptionFn).
func frenchIsException(w string) bool {
	// remove: qq, qqe...
	if strings.HasPrefix(w, "qq") {
		return true
	}
	// informè V ind pres 1 s
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

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer + isException filter).
func (s *FrenchSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer + isException filter).
func (s *FrenchSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*FrenchSynthesizer)(nil)
