package en

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// OrdinalSuffixFilter ports org.languagetool.rules.en.OrdinalSuffixFilter.
// Fixes the suggestion for ordinal suffixes, like "1nd" -> "1st".
// Java @since 5.3
type OrdinalSuffixFilter struct{}

func NewOrdinalSuffixFilter() *OrdinalSuffixFilter {
	return &OrdinalSuffixFilter{}
}

// Java:
//   private static final Pattern PATTERN = Pattern.compile(".*(11|12|13)");
//   private static final Pattern NON_DIGITS = Pattern.compile("[^0-9]");
// PATTERN.matcher(ordinal).matches() is full-string match — Go MatchString is
// substring search, so anchors emulate Java Matcher.matches().
var (
	ordinalSuffixTeens     = regexp.MustCompile(`\A.*(11|12|13)\z`)
	ordinalSuffixNonDigits = regexp.MustCompile(`[^0-9]`)
)

// AcceptRuleMatch ports OrdinalSuffixFilter.acceptRuleMatch.
// Java:
//
//	String ordinal = NON_DIGITS.matcher(match.getSuggestedReplacements().get(0)).replaceAll("");
//	if (PATTERN.matcher(ordinal).matches()) {
//	  match.setSuggestedReplacement(ordinal + "th");
//	} else if (ordinal.endsWith("1")) {
//	  match.setSuggestedReplacement(ordinal + "st");
//	} else if (ordinal.endsWith("2")) {
//	  match.setSuggestedReplacement(ordinal + "nd");
//	} else if (ordinal.endsWith("3")) {
//	  match.setSuggestedReplacement(ordinal + "rd");
//	} else {
//	  match.setSuggestedReplacement(ordinal + "th");
//	}
//	return match;
//
// Empty suggestions: Java get(0) → IndexOutOfBoundsException; Go index panic.
func (f *OrdinalSuffixFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	// Java: match.getSuggestedReplacements().get(0) — panics if empty / nil list path.
	ordinal := ordinalSuffixNonDigits.ReplaceAllString(match.GetSuggestedReplacements()[0], "")
	if ordinalSuffixTeens.MatchString(ordinal) {
		match.SetSuggestedReplacement(ordinal + "th")
	} else if strings.HasSuffix(ordinal, "1") {
		match.SetSuggestedReplacement(ordinal + "st")
	} else if strings.HasSuffix(ordinal, "2") {
		match.SetSuggestedReplacement(ordinal + "nd")
	} else if strings.HasSuffix(ordinal, "3") {
		match.SetSuggestedReplacement(ordinal + "rd")
	} else {
		match.SetSuggestedReplacement(ordinal + "th")
	}
	return match
}
