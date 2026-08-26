package filters

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// ArabicDMYDateCheckFilter ports org.languagetool.rules.ar.filters.ArabicDMYDateCheckFilter.
// Date filter that expects a 'date' argument in the format 'dd-mm-yyyy'.
type ArabicDMYDateCheckFilter struct {
	*ArabicDateCheckFilter
}

func NewArabicDMYDateCheckFilter() *ArabicDMYDateCheckFilter {
	return &ArabicDMYDateCheckFilter{ArabicDateCheckFilter: NewArabicDateCheckFilter()}
}

// AcceptRuleMatch ports ArabicDMYDateCheckFilter.acceptRuleMatch.
//
//	if (args.containsKey("year") || args.containsKey("month") || args.containsKey("day")) {
//	  throw new RuntimeException("Set only 'weekDay' and 'date' for " + ArabicDMYDateCheckFilter.class.getSimpleName());
//	}
//	String dateString = getRequired("date", args);
//	String[] parts = dateString.split("-");
//	if (parts.length != 3) {
//	  throw new RuntimeException("Expected date in format 'dd-mm-yyyy': '" + dateString + "'");
//	}
//	args.put("day", parts[0]);
//	args.put("month", parts[1]);
//	args.put("year", parts[2]);
//	return super.acceptRuleMatch(...);
func (f *ArabicDMYDateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	// Java: args.containsKey — key present even if value is empty.
	if _, ok := args["year"]; ok {
		panic("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		panic("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		panic("Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter")
	}
	// Java RuleFilter.getRequired → IllegalArgumentException "Missing key '…'"
	dateString := patterns.GetRequired("date", args)
	// Java String.split("-") default limit 0: trailing empty segments discarded.
	// No delimiter match (e.g. "") → one-element array of the input (Java early return).
	parts := javaSplitDashLimit0(dateString)
	if len(parts) != 3 {
		panic("Expected date in format 'dd-mm-yyyy': '" + dateString + "'")
	}
	// Java mutates args map in place then super.acceptRuleMatch.
	args["day"] = parts[0]
	args["month"] = parts[1]
	args["year"] = parts[2]
	return f.ArabicDateCheckFilter.AcceptRuleMatch(match, args, patternTokenPos, patternTokens, tokenPositions)
}

// javaSplitDashLimit0 ports Java String.split("-") with default limit 0
// (trailing empty strings discarded; if the delimiter never matches, returns {input}).
func javaSplitDashLimit0(s string) []string {
	if !strings.Contains(s, "-") {
		return []string{s}
	}
	parts := strings.Split(s, "-")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
