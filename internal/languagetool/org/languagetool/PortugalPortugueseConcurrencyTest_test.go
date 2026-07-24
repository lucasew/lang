package languagetool

// Twin of PortugalPortugueseConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of PortugalPortugueseConcurrencyTest (Java @Ignore slow spell race)
func TestPortugalPortugueseConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "pt-PT", "Um teste.")
}
