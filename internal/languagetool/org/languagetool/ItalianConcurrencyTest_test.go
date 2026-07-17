package languagetool

// Twin of ItalianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of ItalianConcurrencyTest (Java @Ignore slow spell race)
func TestItalianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "it", "Un test.")
}
