package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// EnglishNgramProbabilityRule ports org.languagetool.rules.en.EnglishNgramProbabilityRule.
type EnglishNgramProbabilityRule struct {
	*ngrams.NgramProbabilityRule
	DefaultOff bool
}

func NewEnglishNgramProbabilityRule(lm ngrams.LanguageModel) *EnglishNgramProbabilityRule {
	return &EnglishNgramProbabilityRule{
		NgramProbabilityRule: ngrams.NewNgramProbabilityRule(lm),
		DefaultOff:           true,
	}
}
