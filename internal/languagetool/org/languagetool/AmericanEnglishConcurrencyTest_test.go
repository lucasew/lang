package languagetool

// Twin of AmericanEnglishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of AmericanEnglishConcurrencyTest (Java @Ignore slow spell race)
func TestAmericanEnglishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "en-US", "A test.")
}
