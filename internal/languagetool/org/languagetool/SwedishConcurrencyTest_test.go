package languagetool

// Twin of SwedishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of SwedishConcurrencyTest (Java @Ignore slow spell race)
func TestSwedishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "sv", "Ett test.")
}
