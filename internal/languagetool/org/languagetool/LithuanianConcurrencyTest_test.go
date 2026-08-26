package languagetool

// Twin of LithuanianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of LithuanianConcurrencyTest (Java @Ignore slow spell race)
func TestLithuanianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "lt", "Testas.")
}
