package languagetool

// Twin of GermanyGermanConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of GermanyGermanConcurrencyTest (Java @Ignore slow spell race)
func TestGermanyGermanConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "de-DE", "Ein Test.")
}
