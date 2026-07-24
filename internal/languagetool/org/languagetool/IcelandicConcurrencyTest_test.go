package languagetool

// Twin of IcelandicConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of IcelandicConcurrencyTest (Java @Ignore slow spell race)
func TestIcelandicConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "is", "Próf.")
}
