package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// AgreementRuleAntiPatterns3 holds more DE agreement anti-patterns (subset).
var AgreementRuleAntiPatterns3 = [][]*patterns.PatternToken{
	{
		patterns.Token("pro"),
		patterns.PosRegex("SUB:.*"),
	},
	{
		patterns.TokenRegex("je|pro"),
		patterns.TokenRegex("\\d+"),
		patterns.PosRegex("SUB:.*"),
	},
}
