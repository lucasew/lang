package languagetool

// Twin of languagetool-language-modules/ast/src/test/java/org/languagetool/AsturianConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of AsturianConcurrencyTest (createLanguage=Asturian, createSampleText=…).
func TestAsturianConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "L'asturianu ye una llingua romancep ropia d'Asturies, perteneciente al subgrupu asturllionés."
	ConcurrencyAnalyzeSmoke(t, "ast",
		"L'asturianu ye una llingua romancep ropia d'Asturies, perteneciente al subgrupu asturllionés.")
}
