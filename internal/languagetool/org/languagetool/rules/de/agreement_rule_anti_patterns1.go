package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// MonthNamesRegex ports AgreementRuleAntiPatterns1.MONTH_NAMES_REGEX.
const MonthNamesRegex = "Jänner|Januar|Februar|März|April|Mai|Juni|Juli|August|September|Oktober|November|Dezember"

// AgreementRuleAntiPatterns1 is a subset of ANTI_PATTERNS for DE_AGREEMENT.
// Full list is large; patterns are appended as needed for matching coverage.
var AgreementRuleAntiPatterns1 = [][]*patterns.PatternToken{
	{
		patterns.TokenRegex("bring(s?t|en?)"),
		patterns.Token("das"),
		patterns.PosRegex("ADJ:.*"),
		patterns.PosRegex("SUB:.*PLU.*"),
		patterns.Token("mit"),
		patterns.Token("sich"),
	},
	{
		// "den 1. Januar" style month phrases
		patterns.TokenRegex("den|am|im"),
		patterns.TokenRegex("\\d{1,2}\\.?"),
		patterns.TokenRegex(MonthNamesRegex),
	},
}

// AllAgreementAntiPatterns concatenates anti-pattern batches.
func AllAgreementAntiPatterns() [][]*patterns.PatternToken {
	out := make([][]*patterns.PatternToken, 0, len(AgreementRuleAntiPatterns1)+len(AgreementRuleAntiPatterns2)+len(AgreementRuleAntiPatterns3))
	out = append(out, AgreementRuleAntiPatterns1...)
	out = append(out, AgreementRuleAntiPatterns2...)
	out = append(out, AgreementRuleAntiPatterns3...)
	return out
}
