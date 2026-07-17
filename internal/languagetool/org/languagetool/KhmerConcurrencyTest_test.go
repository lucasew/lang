package languagetool

// Twin of KhmerConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of KhmerConcurrencyTest (Java @Ignore slow spell race)
func TestKhmerConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "km", "សាកល្បង។")
}
