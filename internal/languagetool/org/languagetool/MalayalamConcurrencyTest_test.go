package languagetool

// Twin of MalayalamConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of MalayalamConcurrencyTest (Java @Ignore slow spell race)
func TestMalayalamConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "ml", "test")
}
