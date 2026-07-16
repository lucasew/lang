package patterns

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RuleFilterEvaluator ports org.languagetool.rules.patterns.RuleFilterEvaluator.
type RuleFilterEvaluator struct {
	Filter RuleFilter
}

func NewRuleFilterEvaluator(filter RuleFilter) *RuleFilterEvaluator {
	return &RuleFilterEvaluator{Filter: filter}
}

func (e *RuleFilterEvaluator) RunFilter(filterArgs string, ruleMatch *rules.RuleMatch,
	patternTokens []*languagetool.AnalyzedTokenReadings, patternTokenPos int, tokenPositions []int) *rules.RuleMatch {
	args := GetResolvedArguments(filterArgs, patternTokens, patternTokenPos, tokenPositions)
	return e.Filter.AcceptRuleMatch(ruleMatch, args, patternTokenPos, patternTokens, tokenPositions)
}

// GetResolvedArguments resolves backrefs like \1 to pattern token strings.
func GetResolvedArguments(filterArgs string, patternTokens []*languagetool.AnalyzedTokenReadings, patternTokenPos int, tokenPositions []int) map[string]string {
	_ = patternTokenPos
	result := map[string]string{}
	if strings.TrimSpace(filterArgs) == "" {
		return result
	}
	for _, arg := range strings.Fields(filterArgs) {
		delimPos := strings.Index(arg, ":")
		if delimPos == -1 {
			panic(fmt.Sprintf("Invalid syntax for key/value, expected 'key:value', got: '%s'", arg))
		}
		key := arg[:delimPos]
		val := arg[delimPos+1:]
		if strings.HasPrefix(val, `\`) {
			refNumber, err := strconv.Atoi(strings.ReplaceAll(val, `\`, ""))
			if err != nil {
				panic(err)
			}
			if refNumber > len(tokenPositions) {
				panic(fmt.Sprintf("Your reference number %d is bigger than the number of tokens: %d", refNumber, len(tokenPositions)))
			}
			correctedRef := getSkipCorrectedReference(tokenPositions, refNumber)
			if correctedRef >= len(patternTokens) {
				panic(fmt.Sprintf("Your reference number %d is bigger than number of matching tokens: %d", refNumber, len(patternTokens)))
			}
			if _, dup := result[key]; dup {
				panic(fmt.Sprintf("Duplicate key '%s'", key))
			}
			result[key] = patternTokens[correctedRef].GetToken()
		} else {
			result[key] = val
		}
	}
	return result
}

// getSkipCorrectedReference adapts ref numbers when tokens have skip.
func getSkipCorrectedReference(tokenPositions []int, refNumber int) int {
	correctedRef := 0
	i := 0
	for _, tokenPosition := range tokenPositions {
		if i >= refNumber {
			break
		}
		i++
		correctedRef += tokenPosition
	}
	return correctedRef - 1
}
