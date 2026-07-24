package languagetool

// Twin of BretonConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of BretonConcurrencyTest (Java @Ignore slow spell race)
func TestBretonConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "br", "Un test.")
}
