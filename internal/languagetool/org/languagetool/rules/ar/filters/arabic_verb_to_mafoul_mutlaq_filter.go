package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Default verb→masdar map (subset of ArabicVerbToMafoulMutlaqFilter).
var defaultVerb2Masdar = map[string][]string{
	"عَمِلَ":  {"عمل"},
	"أَعْمَلَ": {"إعمال"},
	"عَمَّلَ":  {"تعميل"},
	"عمل":   {"عمل"},
}

// ArabicVerbToMafoulMutlaqFilter ports absolute-object (مفعول مطلق) suggestions.
type ArabicVerbToMafoulMutlaqFilter struct {
	Verb2Masdar map[string][]string
}

func NewArabicVerbToMafoulMutlaqFilter() *ArabicVerbToMafoulMutlaqFilter {
	m := map[string][]string{}
	for k, v := range defaultVerb2Masdar {
		m[k] = append([]string{}, v...)
	}
	return &ArabicVerbToMafoulMutlaqFilter{Verb2Masdar: m}
}

// MasdarsForVerb returns masdar forms for a verb lemma.
func (f *ArabicVerbToMafoulMutlaqFilter) MasdarsForVerb(verbLemma string) []string {
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

// SuggestMafoulMutlaq builds "masdar" suggestions (optionally doubled).
func (f *ArabicVerbToMafoulMutlaqFilter) SuggestMafoulMutlaq(verbLemma string) []string {
	ms := f.MasdarsForVerb(verbLemma)
	var out []string
	for _, m := range ms {
		out = append(out, m)
		// common absolute object pattern: masdar + masdar (simplified)
		out = append(out, m+" "+m)
	}
	return out
}
