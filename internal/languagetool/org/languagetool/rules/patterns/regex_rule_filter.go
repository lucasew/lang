package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegexRuleFilter ports org.languagetool.rules.patterns.RegexRuleFilter.
// groups mirrors Java Matcher: groups[0] full match, groups[1] capture group 1, …
type RegexRuleFilter interface {
	AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string,
		sentence *languagetool.AnalyzedSentence, groups []string) *rules.RuleMatch
}

// RegexRuleFilterFunc adapts a function.
type RegexRuleFilterFunc func(match *rules.RuleMatch, arguments map[string]string,
	sentence *languagetool.AnalyzedSentence, groups []string) *rules.RuleMatch

func (f RegexRuleFilterFunc) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string,
	sentence *languagetool.AnalyzedSentence, groups []string) *rules.RuleMatch {
	return f(match, arguments, sentence, groups)
}
