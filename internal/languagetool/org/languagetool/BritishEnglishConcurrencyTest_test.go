package languagetool

// Twin of BritishEnglishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of BritishEnglishConcurrencyTest (Java @Ignore slow spell race)
func TestBritishEnglishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "en-GB", "A test.")
}
