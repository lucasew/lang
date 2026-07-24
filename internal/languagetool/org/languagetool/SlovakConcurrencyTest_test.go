package languagetool

// Twin of SlovakConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of SlovakConcurrencyTest (Java @Ignore slow spell race)
func TestSlovakConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "sk", "Test.")
}
