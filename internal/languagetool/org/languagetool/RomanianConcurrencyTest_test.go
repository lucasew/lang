package languagetool

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/RomanianConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of RomanianConcurrencyTest (createLanguage=Romanian, createSampleText=…).
func TestRomanianConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Puteți adăuga un articol în lista de articole cerute pentru ca un alt colaborator să-l înceapă."
	ConcurrencyAnalyzeSmoke(t, "ro",
		"Puteți adăuga un articol în lista de articole cerute pentru ca un alt colaborator să-l înceapă.")
}
