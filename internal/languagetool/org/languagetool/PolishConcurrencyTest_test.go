package languagetool

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/PolishConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of PolishConcurrencyTest (createLanguage=Polish, createSampleText=Wiedźmin…).
func TestPolishConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Wiedźmin – postać stworzona przez polskiego pisarza fantasy Andrzeja Sapkowskiego."
	ConcurrencyAnalyzeSmoke(t, "pl",
		"Wiedźmin – postać stworzona przez polskiego pisarza fantasy Andrzeja Sapkowskiego.")
}
