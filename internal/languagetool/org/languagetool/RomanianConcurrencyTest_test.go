package languagetool

// Twin of RomanianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of RomanianConcurrencyTest (Java @Ignore slow spell race)
func TestRomanianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "ro", "Un test.")
}
