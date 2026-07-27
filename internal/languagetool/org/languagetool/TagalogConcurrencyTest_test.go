package languagetool

// Twin of languagetool-language-modules/tl/src/test/java/org/languagetool/TagalogConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of TagalogConcurrencyTest (createLanguage=Tagalog, createSampleText=…).
func TestTagalogConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Mula sa Wikipediang Tagalog, ang malayang ensiklopedya"
	ConcurrencyAnalyzeSmoke(t, "tl",
		"Mula sa Wikipediang Tagalog, ang malayang ensiklopedya")
}
