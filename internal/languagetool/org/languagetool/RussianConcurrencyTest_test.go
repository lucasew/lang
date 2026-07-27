package languagetool

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/RussianConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of RussianConcurrencyTest (createLanguage=Russian, createSampleText=…).
func TestRussianConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Материал из Википедии — свободной энциклопедии"
	ConcurrencyAnalyzeSmoke(t, "ru",
		"Материал из Википедии — свободной энциклопедии")
}
