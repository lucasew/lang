package pt

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// YMDDateCheckFilter ports org.languagetool.rules.pt.YMDDateCheckFilter.
// Expects 'date' in 'yyyy-mm-dd' form; delegates to DateCheckFilter (suggestions super).
type YMDDateCheckFilter struct {
	dateCheck *DateCheckFilter
	ymd       *rules.YMDDateHelper
}

func NewYMDDateCheckFilter() *YMDDateCheckFilter {
	return &YMDDateCheckFilter{
		dateCheck: NewDateCheckFilter(),
		ymd:       rules.NewYMDDateHelper(),
	}
}

// PrepareArgs parses date=yyyy-mm-dd and rejects year/month/day keys (test helper).
func (f *YMDDateCheckFilter) PrepareArgs(args map[string]string) (map[string]string, error) {
	if _, ok := args["year"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if f == nil || f.ymd == nil {
		return nil, fmt.Errorf("missing key 'date'")
	}
	return f.ymd.ParseDate(args)
}

// AcceptRuleMatch ports YMDDateCheckFilter.acceptRuleMatch.
// Java: reject year/month/day keys; parseDate(args); super.acceptRuleMatch(...).
func (f *YMDDateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if _, ok := args["year"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if f.ymd == nil || f.dateCheck == nil {
		return nil
	}
	parsed, err := f.ymd.ParseDate(args)
	if err != nil {
		panic(err.Error())
	}
	return f.dateCheck.AcceptRuleMatch(match, parsed, 0, patternTokens, tokenPositions)
}
