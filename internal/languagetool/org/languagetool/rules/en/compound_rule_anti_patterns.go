package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// CompoundRuleAntiPatterns ports CompoundRule.ANTI_PATTERNS (16/16).
// Java: makeAntiPatterns(..., AMERICAN_ENGLISH) → IMMUNIZE DisambiguationPatternRules.
// POS-gated arms (VB.*, VBG, IN|TO, SENT_START|CC|PCT) only fire when tokens
// carry those POS tags; surface-only arms work without a tagger.
var CompoundRuleAntiPatterns = [][]*patterns.PatternToken{
	// ['’`´‘] re  — use double-quoted regex: raw string cannot contain backtick
	{
		patterns.TokenRegex("['\u2019`\u00b4\u2018]"),
		patterns.Token("re"),
	},
	// We well received … — SENT_START|CC|PCT + pronoun + well + VB.*
	{
		patterns.PosRegex("SENT_START|CC|PCT"),
		patterns.TokenRegex("we|you|they|I|s?he|it"),
		patterns.Token("well"),
		patterns.PosRegex("VB.*"),
	},
	// how well VB.*
	{
		patterns.Token("how"),
		patterns.Token("well"),
		patterns.PosRegex("VB.*"),
	},
	// and|& co
	{
		patterns.TokenRegex("and|&"),
		patterns.Token("co"),
	},
	// power off key
	{
		patterns.Token("power"),
		patterns.Token("off"),
		patterns.Token("key"),
	},
	// see saw seen
	{
		patterns.Token("see"),
		patterns.Token("saw"),
		patterns.Token("seen"),
	},
	// forward looking IN|TO
	{
		patterns.Token("forward"),
		patterns.Token("looking"),
		patterns.PosRegex("IN|TO"),
	},
	// store front doors?
	{
		patterns.Token("store"),
		patterns.Token("front"),
		patterns.TokenRegex("doors?"),
	},
	// from surface to surface
	{
		patterns.Token("from"),
		patterns.Token("surface"),
		patterns.Token("to"),
		patterns.Token("surface"),
	},
	// senior|junior year end
	{
		patterns.TokenRegex("senior|junior"),
		patterns.Token("year"),
		patterns.Token("end"),
	},
	// under investment banking
	{
		patterns.Token("under"),
		patterns.Token("investment"),
		patterns.Token("banking"),
	},
	// spring cleans?|cleaned|cleaning up|the|my|our|his|her
	{
		patterns.Token("spring"),
		patterns.TokenRegex("cleans?|cleaned|cleaning"),
		patterns.TokenRegex("up|the|my|our|his|her"),
	},
	// series? a  (Serie A / A-Team)
	{
		patterns.TokenRegex("series?"),
		patterns.TokenRegex("a"),
	},
	// hard time VBG
	{
		patterns.Token("hard"),
		patterns.Token("time"),
		patterns.PosRegex("VBG"),
	},
	// first ever green
	{
		patterns.Token("first"),
		patterns.TokenRegex("ever"),
		patterns.TokenRegex("green"),
	},
	// inter-state.com — .+ . (com|io|…)
	{
		patterns.TokenRegex(".+"),
		patterns.Token("."),
		patterns.TokenRegex("(com|io|de|nl|co|net|org|es)"),
	},
}
