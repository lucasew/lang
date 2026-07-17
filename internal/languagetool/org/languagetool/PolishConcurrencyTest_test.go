package languagetool

// Twin of PolishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of PolishConcurrencyTest (Java @Ignore slow spell race)
func TestPolishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "pl", "Test.")
}
