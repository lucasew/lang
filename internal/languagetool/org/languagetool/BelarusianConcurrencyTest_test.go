package languagetool

// Twin of BelarusianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of BelarusianConcurrencyTest (Java @Ignore slow spell race)
func TestBelarusianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "be", "Тэст.")
}
