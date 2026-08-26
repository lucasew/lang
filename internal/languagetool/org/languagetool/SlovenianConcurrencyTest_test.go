package languagetool

// Twin of languagetool-language-modules/sl/src/test/java/org/languagetool/SlovenianConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of SlovenianConcurrencyTest (createLanguage=Slovenian, createSampleText=…).
func TestSlovenianConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Iz Wikipedije, proste enciklopedije"
	ConcurrencyAnalyzeSmoke(t, "sl",
		"Iz Wikipedije, proste enciklopedije")
}
