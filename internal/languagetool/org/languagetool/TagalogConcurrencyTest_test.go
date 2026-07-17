package languagetool

// Twin of TagalogConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of TagalogConcurrencyTest (Java @Ignore slow spell race)
func TestTagalogConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "tl", "Isang pagsusulit.")
}
