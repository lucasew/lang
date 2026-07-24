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
	sentence *languagetool.AnalyzedSentence, groups []string) *rules.RuleMatch {
	args := ResolveFilterArguments(filterArgs)
	return e.Filter.AcceptRuleMatch(ruleMatch, args, sentence, groups)
}

// ResolveFilterArguments ports RegexRuleFilterEvaluator: filterArgs.split("\\s+").
// Java Pattern \\s without UNICODE_CHARACTER_CLASS (ASCII whitespace only).
func ResolveFilterArguments(filterArgs string) map[string]string {
	result := map[string]string{}
	// Java "".split("\\s+") → {""}; skip empty segments like non-empty split pieces.
	for _, arg := range whitespaceSplit(filterArgs) {
		if arg == "" {
			continue
		}
		delimPos := strings.Index(arg, ":")
		if delimPos == -1 {
			panic(fmt.Sprintf("Invalid syntax for key/value, expected 'key:value', got: '%s'", arg))
		}
		result[arg[:delimPos]] = arg[delimPos+1:]
	}
	return result
}
