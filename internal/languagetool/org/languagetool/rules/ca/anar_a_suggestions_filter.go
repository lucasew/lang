package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AnarASuggestionsFilter ports suggestion assembly for ANAR_A_INFINITIU.
// Synthesizer is pluggable (future + present of infinitive lemma).
type AnarASuggestionsFilter struct {
	// SynthFuturePresent returns future then present forms for the infinitive lemma
	// given the person/number suffix from the "anar" form (e.g. "1S0").
	SynthFuturePresent func(lemma, personNumberSuffix string) []string
}

func NewAnarASuggestionsFilter() *AnarASuggestionsFilter {
	return &AnarASuggestionsFilter{}
}

// Suggest builds "li ho farem / li ho fem" style replacements.
// personNumberSuffix is anarPostag[4:8] (e.g. "1S0."); pronouns is the after/before clitic string.
func (f *AnarASuggestionsFilter) Suggest(lemma, personNumberSuffix, pronouns, casingModel string) []string {
	if f.SynthFuturePresent == nil {
		return nil
	}
	forms := f.SynthFuturePresent(lemma, personNumberSuffix)
	if len(forms) == 0 {
		return nil
	}
	var out []string
	for _, verb := range forms {
		s := ""
		if pronouns != "" {
			s = TransformDavant(pronouns, verb)
		}
		s += verb
		if casingModel != "" {
			s = tools.PreserveCase(s, casingModel)
		}
		// also run AdaptSuggestion
		s = AdaptSuggestion(s, "")
		out = append(out, s)
	}
	return out
}
