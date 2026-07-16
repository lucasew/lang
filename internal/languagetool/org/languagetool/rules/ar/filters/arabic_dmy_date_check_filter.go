package filters

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicDMYDateCheckFilter ports the dd-mm-yyyy date argument form.
// (Named type already introduced earlier; this file provides the real ParseDMY + Accept.)
func ParseDMYDateArg(dateString string) (day, month, year string, err error) {
	parts := strings.Split(dateString, "-")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("expected date in format 'dd-mm-yyyy': %q", dateString)
	}
	return parts[0], parts[1], parts[2], nil
}

// AcceptDMYRuleMatch expands 'date' into day/month/year then runs the abstract filter.
func (f *ArabicDMYDateCheckFilter) AcceptDMYRuleMatch(match *rules.RuleMatch, args map[string]string) (*rules.RuleMatch, error) {
	if f == nil {
		return nil, nil
	}
	if _, ok := args["year"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	dateString, ok := args["date"]
	if !ok || dateString == "" {
		return nil, fmt.Errorf("missing required argument 'date'")
	}
	d, m, y, err := ParseDMYDateArg(dateString)
	if err != nil {
		return nil, err
	}
	// copy args
	expanded := map[string]string{}
	for k, v := range args {
		expanded[k] = v
	}
	expanded["day"] = d
	expanded["month"] = m
	expanded["year"] = y
	return f.AcceptRuleMatch(match, expanded), nil
}
