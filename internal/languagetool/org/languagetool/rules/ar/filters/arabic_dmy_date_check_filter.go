package filters

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicDMYDateCheckFilter ports org.languagetool.rules.ar.filters.ArabicDMYDateCheckFilter.
//
// NOT claimed ready under sector docs/faithful-port-checklist.md#3.C.3-ArabicDateCheckFilter.
// That sector covers ArabicDateCheckFilter + ArabicDateFilterHelper only. DMY needs its own
// sector for the Java acceptRuleMatch that expands args key "date" (dd-mm-yyyy).
type ArabicDMYDateCheckFilter struct {
	*ArabicDateCheckFilter
}

func NewArabicDMYDateCheckFilter() *ArabicDMYDateCheckFilter {
	return &ArabicDMYDateCheckFilter{ArabicDateCheckFilter: NewArabicDateCheckFilter()}
}

// ParseDMYDateArg splits 'dd-mm-yyyy' the way Java dateString.split("-") does when length==3.
// DMY sector ownership — not asserted faithful here.
func ParseDMYDateArg(dateString string) (day, month, year string, err error) {
	parts := strings.Split(dateString, "-")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("Expected date in format 'dd-mm-yyyy': '%s'", dateString)
	}
	return parts[0], parts[1], parts[2], nil
}

// AcceptRuleMatch ports ArabicDMYDateCheckFilter.acceptRuleMatch:
// expand date → day/month/year then super.acceptRuleMatch.
// Kept for compile/use wiring; DMY sector must still validate full parity separately.
func (f *ArabicDMYDateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil {
		return nil
	}
	if _, ok := args["year"]; ok {
		panic("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		panic("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		panic("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	dateString, ok := args["date"]
	if !ok {
		panic("Missing key 'date'")
	}
	d, m, y, err := ParseDMYDateArg(dateString)
	if err != nil {
		panic(err.Error())
	}
	// Java mutates args map in place
	args["day"] = d
	args["month"] = m
	args["year"] = y
	return f.ArabicDateCheckFilter.AcceptRuleMatch(match, args, patternTokenPos, patternTokens, tokenPositions)
}

// AcceptDMYRuleMatch is a legacy test entrypoint (returns error instead of panic).
// Prefer AcceptRuleMatch for Java-shaped call sites. Not part of the DateCheck sector claim.
func (f *ArabicDMYDateCheckFilter) AcceptDMYRuleMatch(match *rules.RuleMatch, args map[string]string) (*rules.RuleMatch, error) {
	if f == nil {
		return nil, nil
	}
	if _, ok := args["year"]; ok {
		return nil, fmt.Errorf("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		return nil, fmt.Errorf("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		return nil, fmt.Errorf("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	dateString, ok := args["date"]
	if !ok || dateString == "" {
		return nil, fmt.Errorf("Missing key 'date'")
	}
	d, m, y, err := ParseDMYDateArg(dateString)
	if err != nil {
		return nil, err
	}
	expanded := map[string]string{}
	for k, v := range args {
		expanded[k] = v
	}
	expanded["day"] = d
	expanded["month"] = m
	expanded["year"] = y
	return f.ArabicDateCheckFilter.AcceptRuleMatch(match, expanded, 0, nil, nil), nil
}
