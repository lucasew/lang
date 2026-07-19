package en

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// OrdinalSuffixFilter ports org.languagetool.rules.en.OrdinalSuffixFilter.
// Fixes ordinal suggestions like "1nd" → "1st".
type OrdinalSuffixFilter struct{}

func NewOrdinalSuffixFilter() *OrdinalSuffixFilter {
	return &OrdinalSuffixFilter{}
}

// Java: Pattern.compile(".*(11|12|13)") and Pattern.compile("[^0-9]")
var (
	ordinalTeens    = regexp.MustCompile(`.*(11|12|13)`)
	ordinalNonDigit = regexp.MustCompile(`[^0-9]`)
)

// Fix returns the corrected ordinal string from a broken suggestion (digits extracted).
func (f *OrdinalSuffixFilter) Fix(suggestion string) string {
	ordinal := ordinalNonDigit.ReplaceAllString(suggestion, "")
	if ordinalTeens.MatchString(ordinal) {
		return ordinal + "th"
	}
	switch {
	case strings.HasSuffix(ordinal, "1"):
		return ordinal + "st"
	case strings.HasSuffix(ordinal, "2"):
		return ordinal + "nd"
	case strings.HasSuffix(ordinal, "3"):
		return ordinal + "rd"
	default:
		return ordinal + "th"
	}
}

// AcceptRuleMatch ports OrdinalSuffixFilter.acceptRuleMatch.
// Java: rewrites first suggested replacement ordinal suffix in place.
func (f *OrdinalSuffixFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	reps := match.GetSuggestedReplacements()
	if len(reps) == 0 {
		// Java: getSuggestedReplacements().get(0) → IndexOutOfBoundsException
		panic("OrdinalSuffixFilter: no suggested replacements")
	}
	match.SetSuggestedReplacement(f.Fix(reps[0]))
	return match
}
