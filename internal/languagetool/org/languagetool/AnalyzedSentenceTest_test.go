package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/AnalyzedSentenceTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/AnalyzedSentenceTest.java :: AnalyzedSentenceTest.testToString
func TestAnalyzedSentence_ToString(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-core/src/test/java/org/languagetool/AnalyzedSentenceTest.java :: AnalyzedSentenceTest.testCopy
func TestAnalyzedSentence_Copy(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}
