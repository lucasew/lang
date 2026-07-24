package es

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// SpanishRepeatedWordsAntiPatterns ports SpanishRepeatedWordsRule.ANTI_PATTERNS (5/5).
var SpanishRepeatedWordsAntiPatterns = [][]*patterns.PatternToken{
	// también …
	{patterns.Token("también"), patterns.CsRegex(".+")},
	// … también
	{patterns.CsRegex(".+"), patterns.Token("también")},
	// Antes|Después de|del
	{patterns.CsRegex("[Aa]ntes|[Dd]espués"), patterns.CsRegex("de|del")},
	// Tema N / roman
	{patterns.CsRegex("[Tt]ema|TEMA"), patterns.CsRegex(`\d+|[IXVC]+`)},
	// Así que
	{patterns.CsRegex("[Aa]sí"), patterns.Token("que")},
}
