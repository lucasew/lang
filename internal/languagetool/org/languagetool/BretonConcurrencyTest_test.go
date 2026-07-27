package languagetool

// Twin of languagetool-language-modules/br/src/test/java/org/languagetool/BretonConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of BretonConcurrencyTest (createLanguage=Breton, createSampleText=…).
func TestBretonConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Ma oa bet krouet ar pennad-mañ ganeoc'h pellik zo, marteze eo bet diverketabaoe."
	ConcurrencyAnalyzeSmoke(t, "br",
		"Ma oa bet krouet ar pennad-mañ ganeoc'h pellik zo, marteze eo bet diverketabaoe.")
}
