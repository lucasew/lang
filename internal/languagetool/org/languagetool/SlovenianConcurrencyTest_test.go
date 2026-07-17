package languagetool

// Twin of SlovenianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of SlovenianConcurrencyTest (Java @Ignore slow spell race)
func TestSlovenianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "sl", "Preizkus.")
}
