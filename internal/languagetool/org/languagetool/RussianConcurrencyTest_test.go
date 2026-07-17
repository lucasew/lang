package languagetool

// Twin of RussianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of RussianConcurrencyTest (Java @Ignore slow spell race)
func TestRussianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "ru", "Тест.")
}
