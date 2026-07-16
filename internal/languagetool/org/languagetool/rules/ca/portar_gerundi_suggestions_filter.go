package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// PortarGerundiSuggestionsFilter ports suggestion assembly for PORTAR_GERUNDI.
// Synthesis of "haver + participle" / finite forms is pluggable.
type PortarGerundiSuggestionsFilter struct {
	// SynthHaverParticiple returns "he fet"-style candidates for the portar person/number.
	SynthHaverParticiple func(lemma, portarPostagSuffix string) []string
	// SynthFinite returns finite forms of the gerund lemma matching portar tense/person.
	SynthFinite func(lemma, portarPostagSuffix string) []string
}

func NewPortarGerundiSuggestionsFilter() *PortarGerundiSuggestionsFilter {
	return &PortarGerundiSuggestionsFilter{}
}

// Suggest builds replacements for "porto fent-ho" style matches.
// portarPostag is the full V.[IS]… tag of "portar"; lemma is the gerund lemma.
// pronounsAfter is the weak-pronoun cluster after the gerund (may be empty).
// casingModel is the original "portar" token for PreserveCase.
func (f *PortarGerundiSuggestionsFilter) Suggest(portarPostag, lemma, pronounsAfter, casingModel string) []string {
	if len(portarPostag) < 8 {
		return nil
	}
	suffix := portarPostag[2:] // from mood/tense onward as Java VA/V. + substring(2)
	var raw []string
	if f.SynthHaverParticiple != nil {
		// Java: VA + atr1.postag.substring(2)
		raw = append(raw, f.SynthHaverParticiple(lemma, suffix)...)
	}
	if f.SynthFinite != nil {
		// Java: V. + atr1.postag.substring(2)
		if forms := f.SynthFinite(lemma, suffix); len(forms) > 0 {
			raw = append(raw, forms[0])
		}
	}
	if len(raw) == 0 {
		return nil
	}
	var out []string
	for _, r := range raw {
		s := r
		if pronounsAfter != "" {
			s = TransformDavant(pronounsAfter, r) + r
		}
		if casingModel != "" {
			s = tools.PreserveCase(s, casingModel)
		}
		out = append(out, s)
	}
	return out
}

// JoinHaverParticiple is a helper for tests building "he fet" pairs.
func JoinHaverParticiple(haverForms, partForms []string) []string {
	var out []string
	for _, h := range haverForms {
		for _, p := range partForms {
			out = append(out, h+" "+p)
		}
	}
	return out
}
