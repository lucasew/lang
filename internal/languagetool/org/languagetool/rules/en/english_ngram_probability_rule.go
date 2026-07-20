package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// EnglishNgramProbabilityRule ports org.languagetool.rules.en.EnglishNgramProbabilityRule.
type EnglishNgramProbabilityRule struct {
	*ngrams.NgramProbabilityRule
	DefaultOff bool
}

func NewEnglishNgramProbabilityRule(lm ngrams.LanguageModel) *EnglishNgramProbabilityRule {
	base := ngrams.NewNgramProbabilityRule(lm)
	// Java getGoogleStyleWordTokenizer → GoogleStyleWordTokenizer
	gst := NewGoogleStyleWordTokenizer()
	base.Tokenize = gst.Tokenize
	return &EnglishNgramProbabilityRule{
		NgramProbabilityRule: base,
		// Java setDefaultOff() — too many false alarms (2015-12)
		DefaultOff: true,
	}
}

// IsDefaultOff ports Rule.isDefaultOff.
func (r *EnglishNgramProbabilityRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}
