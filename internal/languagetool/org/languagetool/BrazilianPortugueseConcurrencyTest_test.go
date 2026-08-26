package languagetool

// Twin of BrazilianPortugueseConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of BrazilianPortugueseConcurrencyTest (Java @Ignore slow spell race)
func TestBrazilianPortugueseConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "pt-BR", "Um teste.")
}
