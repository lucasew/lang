package languagetool

// Twin of ChineseConcurrencyTest — concurrent Analyze smoke (full spell race deferred).
import "testing"

// Port of ChineseConcurrencyTest (Java @Ignore slow spell race)
func TestChineseConcurrency_NoTests(t *testing.T) {
	ConcurrencyAnalyzeSmoke(t, "zh", "测试。")
}
