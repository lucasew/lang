package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// CatalanRepeatedWordsAntiPatterns ports CatalanRepeatedWordsRule.ANTI_PATTERNS (1/1).
var CatalanRepeatedWordsAntiPatterns = [][]*patterns.PatternToken{
	// Tema N / roman
	{patterns.CsRegex("[Tt]ema|TEMA"), patterns.CsRegex(`\d+|[IXVC]+`)},
}
