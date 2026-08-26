package languagetool

// Twin of languagetool-language-modules/da/src/test/java/org/languagetool/DanishConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of DanishConcurrencyTest (createLanguage=Danish, createSampleText=…).
func TestDanishConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Se hjælpesiderne for dokumentation eller find svar på ofte stillede spørgsmål."
	ConcurrencyAnalyzeSmoke(t, "da",
		"Se hjælpesiderne for dokumentation eller find svar på ofte stillede spørgsmål.")
}
