package languagetool

// Twin of AustrianGermanConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of AustrianGermanConcurrencyTest (Java @Ignore slow spell race)
func TestAustrianGermanConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "de-AT", "Ein Test.")
}
