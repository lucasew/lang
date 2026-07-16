package rules

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractDateCheckWithSuggestionsFilter ports
// org.languagetool.rules.AbstractDateCheckWithSuggestionsFilter.
// Resolves weekDay/day/month/year (or date=yyyy-mm-dd) from pattern tokens.
type AbstractDateCheckWithSuggestionsFilter struct {
	AbstractDateCheckFilter
	// ErrorMessageWrongYear optional message when year/date is invalid.
	ErrorMessageWrongYear string
}

// AcceptRuleMatch resolves args as 1-based token indexes (with skip correction)
// or keeps literal year/month/day/weekDay strings when not numeric indexes.
func (f *AbstractDateCheckWithSuggestionsFilter) AcceptRuleMatch(
	match *RuleMatch, args map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int,
) *RuleMatch {
	if f == nil {
		return nil
	}
	resolved := map[string]string{}
	for k, v := range args {
		resolved[k] = v
	}
	if dateIdxStr, ok := args["date"]; ok && dateIdxStr != "" && dateIdxStr != "-1" {
		if idx, err := strconv.Atoi(dateIdxStr); err == nil {
			ref := skipCorrectedRef(tokenPositions, idx)
			if ref >= 0 && ref < len(patternTokens) {
				parts := strings.Split(patternTokens[ref].GetToken(), "-")
				if len(parts) == 3 {
					resolved["year"] = parts[0]
					resolved["month"] = parts[1]
					resolved["day"] = parts[2]
				}
			}
		}
	}
	if wd, ok := args["weekDay"]; ok {
		if idx, err := strconv.Atoi(wd); err == nil && patternTokens != nil {
			ref := skipCorrectedRef(tokenPositions, idx)
			if ref >= 0 && ref < len(patternTokens) {
				resolved["weekDay"] = strings.ReplaceAll(patternTokens[ref].GetToken(), "\u00AD", "")
			}
		}
	}
	// Only rewrite day/month/year from token indexes when not using full date token.
	if args["date"] == "" || args["date"] == "-1" {
		for _, key := range []string{"day", "month", "year"} {
			if v, ok := args[key]; ok {
				if idx, err := strconv.Atoi(v); err == nil && patternTokens != nil {
					ref := skipCorrectedRef(tokenPositions, idx)
					if ref >= 0 && ref < len(patternTokens) {
						resolved[key] = patternTokens[ref].GetToken()
					}
				}
			}
		}
	}
	out := f.AbstractDateCheckFilter.AcceptRuleMatch(match, resolved)
	if out != nil {
		return out
	}
	if f.ErrorMessageWrongYear != "" {
		if _, err := f.parseDate(resolved); err != nil {
			return NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), f.ErrorMessageWrongYear)
		}
	}
	return nil
}
