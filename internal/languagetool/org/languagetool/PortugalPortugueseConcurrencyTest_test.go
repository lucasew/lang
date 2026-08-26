package languagetool

// Twin of PortugalPortugueseConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of PortugalPortugueseConcurrencyTest (Java @Ignore slow spell race)
func TestPortugalPortugueseConcurrency_NoTests(t *testing.T) {
	// Java createSampleText:
	// "Outras razões pelas quais esta mensagem pode aparecer"
	ConcurrencyAnalyzeSmoke(t, "pt-PT", "Outras razões pelas quais esta mensagem pode aparecer")
}
