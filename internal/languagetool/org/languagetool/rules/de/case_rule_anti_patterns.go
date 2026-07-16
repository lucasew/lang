package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// CaseRuleAntiPatterns is a subset of CaseRule ANTI_PATTERNS.
var CaseRuleAntiPatterns = [][]*patterns.PatternToken{
	{
		// "Guten Tag"
		patterns.TokenRegex("Guten|Guten?"),
		patterns.Token("Tag"),
	},
	{
		patterns.Token("Herr"),
		patterns.PosRegex("EIG:.*"),
	},
	{
		patterns.Token("Frau"),
		patterns.PosRegex("EIG:.*"),
	},
	{
		// month after day
		patterns.TokenRegex("\\d{1,2}\\.?"),
		patterns.TokenRegex(MonthNamesRegex),
	},
}

// CaseRuleAntiPatternsCount reports loaded pattern count (for tests / wiring).
func CaseRuleAntiPatternsCount() int { return len(CaseRuleAntiPatterns) }
