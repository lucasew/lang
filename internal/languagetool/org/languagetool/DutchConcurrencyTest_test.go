package languagetool

// Twin of DutchConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of DutchConcurrencyTest (Java @Ignore slow spell race)
func TestDutchConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "nl", "Een test.")
}
