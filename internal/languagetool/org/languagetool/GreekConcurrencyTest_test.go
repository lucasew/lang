package languagetool

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/GreekConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of GreekConcurrencyTest (createLanguage=Greek, createSampleText=…).
func TestGreekConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Δεν υπάρχει αυτή τη στιγμή λήμμα με αυτόν τον τίτλο."
	ConcurrencyAnalyzeSmoke(t, "el",
		"Δεν υπάρχει αυτή τη στιγμή λήμμα με αυτόν τον τίτλο.")
}
