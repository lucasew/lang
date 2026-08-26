package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// YMDNewYearDateFilter ports org.languagetool.rules.en.YMDNewYearDateFilter.
// Expects a 'date' argument in 'yyyy-mm-dd' form; expands via YMDDateHelper then NewYearDateFilter.
type YMDNewYearDateFilter struct {
	// newYear is the EN NewYearDateFilter super (Java extends NewYearDateFilter).
	newYear *NewYearDateFilter
	ymd     *rules.YMDDateHelper
}

func NewYMDNewYearDateFilter() *YMDNewYearDateFilter {
	return &YMDNewYearDateFilter{
		newYear: NewNewYearDateFilter(),
		ymd:     rules.NewYMDDateHelper(),
	}
}

// AcceptRuleMatch ports YMDNewYearDateFilter.acceptRuleMatch.
// Java:
//
//	if (args.containsKey("year") || args.containsKey("month") || args.containsKey("day")) {
//	  throw new RuntimeException("Set only 'weekDay' and 'date' for " + YMDDateCheckFilter.class.getSimpleName());
//	}
//	args = ymdDateHelper.parseDate(args);
//	return super.acceptRuleMatch(ymdDateHelper.correctDate(match, args), args, patternTokenPos, patternTokens, tokenPositions);
//
// Upstream quirk: exception message uses YMDDateCheckFilter (not YMDNewYearDateFilter).
func (f *YMDNewYearDateFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
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
	// Java: super.acceptRuleMatch(ymdDateHelper.correctDate(match, args), args, ...)
	return f.newYear.AcceptRuleMatch(f.ymd.CorrectDate(match, parsed), parsed, patternTokenPos, patternTokens, tokenPositions)
}
