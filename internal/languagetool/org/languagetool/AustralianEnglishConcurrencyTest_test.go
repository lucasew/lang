package languagetool

// Twin of AustralianEnglishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of AustralianEnglishConcurrencyTest (Java @Ignore slow spell race)
func TestAustralianEnglishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "en-AU", "A test.")
}
