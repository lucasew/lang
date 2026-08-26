package languagetool

// Twin of FrenchConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of FrenchConcurrencyTest (Java @Ignore slow spell race)
func TestFrenchConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "fr", "Un test.")
}
