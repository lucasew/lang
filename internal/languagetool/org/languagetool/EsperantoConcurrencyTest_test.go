package languagetool

// Twin of EsperantoConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of EsperantoConcurrencyTest (Java @Ignore slow spell race)
func TestEsperantoConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "eo", "Testo.")
}
