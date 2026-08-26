package languagetool

// Twin of languagetool-language-modules/sk/src/test/java/org/languagetool/SlovakConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of SlovakConcurrencyTest (createLanguage=Slovak, createSampleText=…).
func TestSlovakConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Wikipédia neobsahuje článok s takýmto názvom."
	ConcurrencyAnalyzeSmoke(t, "sk",
		"Wikipédia neobsahuje článok s takýmto názvom.")
}
