package filters

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// ArabicAdvancedSynthesizerFilter ports org.languagetool.rules.ar.filters.ArabicAdvancedSynthesizerFilter.
type ArabicAdvancedSynthesizerFilter struct {
	*rules.AbstractAdvancedSynthesizerFilter
}

func NewArabicAdvancedSynthesizerFilter(synthesize func(lemma, postag string) []string) *ArabicAdvancedSynthesizerFilter {
	return &ArabicAdvancedSynthesizerFilter{
		AbstractAdvancedSynthesizerFilter: &rules.AbstractAdvancedSynthesizerFilter{
			Synthesize: synthesize,
		},
	}
}
