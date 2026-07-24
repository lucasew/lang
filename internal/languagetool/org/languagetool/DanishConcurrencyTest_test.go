package languagetool

// Twin of DanishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of DanishConcurrencyTest (Java @Ignore slow spell race)
func TestDanishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "da", "En test.")
}
