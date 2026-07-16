package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// AgreementRuleAntiPatterns2 holds additional DE agreement anti-patterns (subset).
var AgreementRuleAntiPatterns2 = [][]*patterns.PatternToken{
	{
		patterns.Token("ein"),
		patterns.TokenRegex("bisschen|wenig|paar"),
	},
	{
		patterns.TokenRegex("viele|manche|einige"),
		patterns.PosRegex("SUB:.*PLU.*"),
	},
}
