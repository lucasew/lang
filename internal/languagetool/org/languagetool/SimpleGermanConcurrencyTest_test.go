package languagetool

// Twin of SimpleGermanConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of SimpleGermanConcurrencyTest (Java @Ignore slow spell race)
func TestSimpleGermanConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "de-DE-x-simple-language", "Ein Test.")
}
