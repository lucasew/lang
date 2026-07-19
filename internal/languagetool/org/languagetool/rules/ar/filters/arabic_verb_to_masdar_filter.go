package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicVerbToMasdarFilter ports verb→masdar lookup from official arabic_verb_masdar.txt
// (Java ArabicVerbToMafoulMutlaqFilter data; not invent reverse of invent maps).
type ArabicVerbToMasdarFilter struct {
	Verb2Masdar map[string][]string
}

func NewArabicVerbToMasdarFilter() *ArabicVerbToMasdarFilter {
	return &ArabicVerbToMasdarFilter{Verb2Masdar: loadOfficialVerbMasdarMap()}
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
