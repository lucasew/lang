package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// QuestionWhitespaceAntiPatterns ports QuestionWhitespaceRule.ANTI_PATTERNS (7/7).
// Java: cacheAntiPatterns(language, ANTI_PATTERNS) → IMMUNIZE DisambiguationPatternRules.
var QuestionWhitespaceAntiPatterns = [][]*patterns.PatternToken{
	// smileys such as :-)
	{
		patterns.TokenRegex("[:;]"),
		patterns.NewPatternTokenBuilder().CsToken("-").SetIsWhiteSpaceBefore(false).Build(),
		patterns.NewPatternTokenBuilder().TokenRegex(`[()D]`).SetIsWhiteSpaceBefore(false).Build(),
	},
	// smileys such as :)
	{
		patterns.TokenRegex("[:;]"),
		patterns.NewPatternTokenBuilder().TokenRegex(`[()D]`).SetIsWhiteSpaceBefore(false).Build(),
	},
	// times like 23:20
	{
		patterns.TokenRegex(`.*\d{1,2}`),
		patterns.Token(":"),
		patterns.TokenRegex(`\d{1,2}`),
	},
	// "??" / "!!"
	{
		patterns.TokenRegex(`[?!]`),
		patterns.TokenRegex(`[?!]`),
	},
	// mac address fragment xx:xx:xx
	{
		patterns.TokenRegex(`[a-z0-9]{2}`),
		patterns.Token(":"),
		patterns.TokenRegex(`[a-z0-9]{2}`),
		patterns.Token(":"),
		patterns.TokenRegex(`[a-z0-9]{2}`),
	},
	// csv markup ;field;
	{
		patterns.Token(";"),
		patterns.NewPatternTokenBuilder().TokenRegex(".+").SetIsWhiteSpaceBefore(false).Build(),
		patterns.NewPatternTokenBuilder().Token(";").SetIsWhiteSpaceBefore(false).Build(),
	},
	// csv markup field;field
	{
		patterns.NewPatternTokenBuilder().TokenRegex(".+").SetIsWhiteSpaceBefore(false).Build(),
		patterns.NewPatternTokenBuilder().Token(";").SetIsWhiteSpaceBefore(false).Build(),
		patterns.NewPatternTokenBuilder().TokenRegex(".+").SetIsWhiteSpaceBefore(false).Build(),
	},
}
