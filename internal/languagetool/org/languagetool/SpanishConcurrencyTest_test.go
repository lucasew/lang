package languagetool

// Twin of SpanishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of SpanishConcurrencyTest (Java @Ignore slow spell race)
func TestSpanishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "es", "Una prueba.")
}
