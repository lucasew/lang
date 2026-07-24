package languagetool

// Twin of SouthAfricanEnglishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of SouthAfricanEnglishConcurrencyTest (Java @Ignore slow spell race)
func TestSouthAfricanEnglishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "en-ZA", "A test.")
}
