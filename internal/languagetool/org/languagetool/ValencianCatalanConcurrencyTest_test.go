package languagetool

// Twin of ValencianCatalanConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of ValencianCatalanConcurrencyTest (Java @Ignore slow spell race)
func TestValencianCatalanConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "ca-ES-valencia", "Una prova.")
}
