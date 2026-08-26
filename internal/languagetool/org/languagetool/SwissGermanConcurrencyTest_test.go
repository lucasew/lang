package languagetool

// Twin of SwissGermanConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of SwissGermanConcurrencyTest (Java @Ignore slow spell race)
func TestSwissGermanConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "de-CH", "Ein Test.")
}
