package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/GlobalSpellingTest.java
// Validation logic and file checks live in rules/spelling (avoid import cycle).
import "testing"

// Port of GlobalSpellingTest.avoidSomeWords — see spelling.TestGlobalSpelling_AvoidSomeWords.
func TestGlobalSpelling_AvoidSomeWords(t *testing.T) {
	t.Log("implemented in rules/spelling.TestGlobalSpelling_AvoidSomeWords")
}
