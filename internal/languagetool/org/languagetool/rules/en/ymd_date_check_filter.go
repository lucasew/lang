package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// YMDDateCheckFilter ports org.languagetool.rules.en.YMDDateCheckFilter.
// Expects a 'date' argument in 'yyyy-mm-dd' form; expands via YMDDateHelper then DateCheckFilter.
type YMDDateCheckFilter struct {
	// dateCheck is the EN DateCheckFilter super (Java extends DateCheckFilter).
	dateCheck *DateCheckFilter
	ymd       *rules.YMDDateHelper
}

func NewYMDDateCheckFilter() *YMDDateCheckFilter {
	return &YMDDateCheckFilter{
		dateCheck: NewDateCheckFilter(),
		ymd:       rules.NewYMDDateHelper(),
	}
}

// AcceptRuleMatch ports YMDDateCheckFilter.acceptRuleMatch.
// Java: if year/month/day keys present → RuntimeException; else parseDate(args); super.acceptRuleMatch(...).
func (f *YMDDateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	// Java: args.containsKey("year") || args.containsKey("month") || args.containsKey("day")
	if _, ok := args["year"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	// Java: ymdDateHelper.parseDate(args) mutates args and may throw IllegalArgumentException / RuntimeException.
	parsed, err := f.ymd.ParseDate(args)
	if err != nil {
		panic(err.Error())
	}
	// Java: super.acceptRuleMatch(match, parsed, patternTokenPos, patternTokens, tokenPositions)
	return f.dateCheck.AcceptRuleMatch(match, parsed, patternTokenPos, patternTokens, tokenPositions)
}
