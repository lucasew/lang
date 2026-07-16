package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegexRuleFilter ports org.languagetool.rules.patterns.RegexRuleFilter.
// patternMatcher is the full match string (Go has no java.util.regex.Matcher state).
type RegexRuleFilter interface {
	AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string,
		sentence *languagetool.AnalyzedSentence, patternMatch string) *rules.RuleMatch
}

// RegexRuleFilterFunc adapts a function.
type RegexRuleFilterFunc func(match *rules.RuleMatch, arguments map[string]string,
	sentence *languagetool.AnalyzedSentence, patternMatch string) *rules.RuleMatch

func (f RegexRuleFilterFunc) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string,
	sentence *languagetool.AnalyzedSentence, patternMatch string) *rules.RuleMatch {
	return f(match, arguments, sentence, patternMatch)
}
