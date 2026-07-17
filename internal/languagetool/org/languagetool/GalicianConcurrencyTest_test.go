package languagetool

// Twin of GalicianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of GalicianConcurrencyTest (Java @Ignore slow spell race)
func TestGalicianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "gl", "Unha proba.")
}
