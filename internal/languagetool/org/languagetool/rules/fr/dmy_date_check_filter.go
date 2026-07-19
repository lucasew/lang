package fr

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DMYDateCheckFilter ports org.languagetool.rules.fr.DMYDateCheckFilter.
// Expects date argument as dd-mm-yyyy; delegates to DateCheckFilter super.
type DMYDateCheckFilter struct {
	dateCheck *DateCheckFilter
}

func NewDMYDateCheckFilter() *DMYDateCheckFilter {
	return &DMYDateCheckFilter{dateCheck: NewDateCheckFilter()}
}

// PrepareArgs expands date=dd-mm-yyyy into day/month/year; rejects direct keys.
func (f *DMYDateCheckFilter) PrepareArgs(args map[string]string) (map[string]string, error) {
	if _, ok := args["year"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	dateString, ok := args["date"]
	if !ok || dateString == "" {
		return nil, fmt.Errorf("missing key 'date'")
	}
	parts := strings.Split(dateString, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected date in format 'dd-mm-yyyy': %q", dateString)
	}
	out := map[string]string{}
	for k, v := range args {
		out[k] = v
	}
	out["day"] = parts[0]
	out["month"] = parts[1]
	out["year"] = parts[2]
	return out, nil
}

// AcceptRuleMatch ports DMYDateCheckFilter.acceptRuleMatch.
// Java: reject year/month/day keys; parse dd-mm-yyyy; super.acceptRuleMatch.
func (f *DMYDateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if _, ok := args["year"]; ok {
		panic("Set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		panic("Set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		panic("Set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	parsed, err := f.PrepareArgs(args)
	if err != nil {
		// Java: getRequired throws / RuntimeException on bad format
		panic(err.Error())
	}
	if f.dateCheck == nil {
		f.dateCheck = NewDateCheckFilter()
	}
	return f.dateCheck.AcceptRuleMatch(match, parsed, 0, patternTokens, tokenPositions)
}
