package languagetool

// Twin of UkrainianConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of UkrainianConcurrencyTest (Java @Ignore slow spell race)
func TestUkrainianConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "uk", "Матеріал з Вікіпедії — вільної енциклопедії.")
}
