package languagetool

// Twin of AsturianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of AsturianConcurrencyTest (Java @Ignore slow spell race)
func TestAsturianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "ast", "Una prueba.")
}
