package languagetool

// Twin of languagetool-language-modules/eo/src/test/java/org/languagetool/EsperantoConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of EsperantoConcurrencyTest (createLanguage=Esperanto, createSampleText=…).
func TestEsperantoConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Kreu la artikolon Enriched text aŭ aldonu peton pri ĝi."
	ConcurrencyAnalyzeSmoke(t, "eo",
		"Kreu la artikolon Enriched text aŭ aldonu peton pri ĝi.")
}
