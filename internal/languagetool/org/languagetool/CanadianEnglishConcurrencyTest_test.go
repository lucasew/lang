package languagetool

// Twin of CanadianEnglishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of CanadianEnglishConcurrencyTest (Java @Ignore slow spell race)
func TestCanadianEnglishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "en-CA", "A test.")
}
