package languagetool

// Twin of languagetool-language-modules/be/src/test/java/org/languagetool/BelarusianConcurrencyTest.java
// extends AbstractLanguageConcurrencyTest — concurrent Analyze with Java sample text.
// Full spell-checker race is @Ignore-class slow in practice; Analyze concurrency
// matches the Go AbstractLanguageConcurrencyTest twin pattern.

import "testing"

// Port of BelarusianConcurrencyTest (createLanguage=Belarusian, createSampleText=…).
func TestBelarusianConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Гэта старонка была сцёрта. Ніжэй паказаны журнал сціранняў і пераносаў для гэтай старонкі."
	ConcurrencyAnalyzeSmoke(t, "be",
		"Гэта старонка была сцёрта. Ніжэй паказаны журнал сціранняў і пераносаў для гэтай старонкі.")
}
