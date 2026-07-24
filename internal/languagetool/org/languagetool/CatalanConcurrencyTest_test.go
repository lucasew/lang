package languagetool

// Twin of CatalanConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of CatalanConcurrencyTest (Java @Ignore slow spell race)
func TestCatalanConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "ca", "Una prova.")
}
