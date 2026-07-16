package filters

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// ArabicVerbToMasdarFilter ports reverse mapping of masdar→verb for suggestions.
type ArabicVerbToMasdarFilter struct {
	Verb2Masdar map[string][]string
}

func NewArabicVerbToMasdarFilter() *ArabicVerbToMasdarFilter {
	// invert default masdar map
	inv := map[string][]string{}
	for masdar, verbs := range defaultMasdar2Verb {
		for _, v := range verbs {
			key := tools.RemoveTashkeel(v)
			inv[key] = append(inv[key], masdar)
			inv[v] = append(inv[v], masdar)
		}
	}
	return &ArabicVerbToMasdarFilter{Verb2Masdar: inv}
}

// SuggestMasdarsForVerb returns masdar lemmas for a verb lemma.
func (f *ArabicVerbToMasdarFilter) SuggestMasdarsForVerb(verbLemma string) []string {
	if f == nil {
		return nil
	}
	if v, ok := f.Verb2Masdar[verbLemma]; ok {
		return append([]string{}, v...)
	}
	if v, ok := f.Verb2Masdar[tools.RemoveTashkeel(verbLemma)]; ok {
		return append([]string{}, v...)
	}
	return nil
}
