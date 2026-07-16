package patterns

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegexRuleFilterEvaluator ports org.languagetool.rules.patterns.RegexRuleFilterEvaluator.
type RegexRuleFilterEvaluator struct {
	Filter RegexRuleFilter
}

func NewRegexRuleFilterEvaluator(filter RegexRuleFilter) *RegexRuleFilterEvaluator {
	return &RegexRuleFilterEvaluator{Filter: filter}
}

func (e *RegexRuleFilterEvaluator) RunFilter(filterArgs string, ruleMatch *rules.RuleMatch,
	sentence *languagetool.AnalyzedSentence, patternMatch string) *rules.RuleMatch {
	args := ResolveFilterArguments(filterArgs)
	return e.Filter.AcceptRuleMatch(ruleMatch, args, sentence, patternMatch)
}

// ResolveFilterArguments parses "key:value key2:value2" space-separated pairs.
func ResolveFilterArguments(filterArgs string) map[string]string {
	result := map[string]string{}
	if strings.TrimSpace(filterArgs) == "" {
		return result
	}
	for _, arg := range strings.Fields(filterArgs) {
		delimPos := strings.Index(arg, ":")
		if delimPos == -1 {
			panic(fmt.Sprintf("Invalid syntax for key/value, expected 'key:value', got: '%s'", arg))
		}
		result[arg[:delimPos]] = arg[delimPos+1:]
	}
	return result
}
