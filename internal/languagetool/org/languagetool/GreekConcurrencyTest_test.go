package languagetool

// Twin of GreekConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of GreekConcurrencyTest (Java @Ignore slow spell race)
func TestGreekConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "el", "Δοκιμή.")
}
