package languagetool

// Twin of NewZealandEnglishConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of NewZealandEnglishConcurrencyTest (Java @Ignore slow spell race)
func TestNewZealandEnglishConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "en-NZ", "A test.")
}
