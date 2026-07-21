package patterns

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WHITESPACE ports RuleFilterEvaluator.WHITESPACE (\\s+).
var whitespaceRE = regexp.MustCompile(`\s+`)

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

// GetResolvedArguments ports RuleFilterEvaluator.getResolvedArguments.
// Resolves backrefs like \1 to pattern token strings; splits args on whitespace.
func GetResolvedArguments(filterArgs string, patternTokens []*languagetool.AnalyzedTokenReadings, patternTokenPos int, tokenPositions []int) map[string]string {
	_ = patternTokenPos
	result := map[string]string{}
	// Java: WHITESPACE.split(filterArgs) — empty string yields one empty element → invalid
	arguments := whitespaceSplit(filterArgs)
	for _, arg := range arguments {
		delimPos := strings.Index(arg, ":")
		if delimPos == -1 {
			panic(fmt.Sprintf("Invalid syntax for key/value, expected 'key:value', got: '%s'", arg))
		}
		// Java: substring(0, delimPos) / substring(delimPos+1) — first ':' only
		key := arg[:delimPos]
		val := arg[delimPos+1:]
		if strings.HasPrefix(val, `\`) {
			// Java: Integer.parseInt(val.replace("\\", "")) — strip all backslashes
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
			// Java duplicate check only on backref branch
			if _, dup := result[key]; dup {
				panic(fmt.Sprintf("Duplicate key '%s'", key))
			}
			result[key] = patternTokens[correctedRef].GetToken()
		} else {
			// literal values: last write wins (no duplicate throw)
			result[key] = val
		}
	}
	return result
}

// whitespaceSplit ports Pattern.compile("\\s+").split(s) for non-empty splits.
// Java split on "" returns [""]; on "a  b" returns ["a","b"].
func whitespaceSplit(s string) []string {
	// Java Pattern.split does not discard trailing empty; leading empty kept for leading space.
	// strings.Fields collapses and trims — wrong for "" and " a". Use Split with filter.
	parts := whitespaceRE.Split(s, -1)
	// Java: for empty input, split returns {""}
	if s == "" {
		return []string{""}
	}
	// drop only pure trailing empties from trailing whitespace? Java "a " → ["a"]
	// Pattern.split discards trailing empty strings by default (limit 0).
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// getSkipCorrectedReference ports RuleFilterEvaluator.getSkipCorrectedReference.
// when there's a 'skip', we need to adapt the reference number.
func getSkipCorrectedReference(tokenPositions []int, refNumber int) int {
	// Java:
	//   int correctedRef = 0; int i = 0;
	//   for (int tokenPosition : tokenPositions) {
	//     if (i++ >= refNumber) break;
	//     correctedRef += tokenPosition;
	//   }
	//   return correctedRef - 1;
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
