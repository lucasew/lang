package patterns

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// RuleMatcher ports org.languagetool.rules.patterns.RuleMatcher.
// Match results are any (typically []*rules.RuleMatch) to avoid import cycles
// when only the matcher interface is needed.
type RuleMatcher interface {
	Match(sentence *languagetool.AnalyzedSentence) (any, error)
}

// RuleMatcherFunc adapts a function to RuleMatcher.
type RuleMatcherFunc func(sentence *languagetool.AnalyzedSentence) (any, error)

func (f RuleMatcherFunc) Match(sentence *languagetool.AnalyzedSentence) (any, error) {
	return f(sentence)
}
