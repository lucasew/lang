package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// GermanCompoundRuleAntiPatterns ports GermanCompoundRule.ANTI_PATTERNS (8/8).
// Java: makeAntiPatterns(..., GermanyGerman) → IMMUNIZE DisambiguationPatternRules.
// SwissCompoundRule extends GermanCompoundRule and inherits the same list.
var GermanCompoundRuleAntiPatterns = [][]*patterns.PatternToken{
	// "Die Bürger konnten an die 900 Meter Kabel in Eigenregie verlegen."
	{
		patterns.TokenRegex("an|um"),
		patterns.Token("die"),
		patterns.TokenRegex(`\d+`),
	},
	// "Lohnt sich die Werbung vom ausgegebenen Euro aus gedacht?"
	{
		patterns.NewPatternTokenBuilder().TokenRegex("von|vom").SetSkip(5).Build(),
		patterns.Token("aus"),
		patterns.Token("gedacht"),
	},
	// "… rund 250 Liter Diesel …"
	{
		patterns.TokenRegex("rund|etwa|zirka|cirka|ungefähr|annähernd|grob|wohl|gegen|schätzungsweise"),
		patterns.TokenRegex(`\d+`),
	},
	// "… ca. 900 Meter …"
	{
		patterns.Token("ca"),
		patterns.Token("."),
		patterns.TokenRegex(`\d+`),
	},
	// Eigenname: Kung Fu Panda|Fighting
	{
		patterns.Token("Kung"),
		patterns.Token("Fu"),
		patterns.TokenRegex("Panda|Fighting"),
	},
	// Eigenname: Harlem Gospel Singers
	{
		patterns.Token("Harlem"),
		patterns.Token("Gospel"),
		patterns.Token("Singers"),
	},
	// Englisch: Always on my|your|the|an?|their
	{
		patterns.Token("Always"),
		patterns.Token("on"),
		patterns.TokenRegex("my|your|the|an?|their"),
	},
	// sich selbst gerecht werden
	{
		patterns.TokenRegex("sich|uns|ihm|ihr|mir|euch"),
		patterns.Token("selbst"),
		patterns.TokenRegex("gerecht.*"),
	},
}
