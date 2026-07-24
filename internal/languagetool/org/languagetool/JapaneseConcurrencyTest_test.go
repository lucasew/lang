package languagetool

// Twin of JapaneseConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of JapaneseConcurrencyTest (Java @Ignore slow spell race)
func TestJapaneseConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "ja", "テスト。")
}
