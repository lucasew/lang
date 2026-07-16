package errorcorpus

// Twin of languagetool-dev/src/test/java/org/languagetool/dev/errorcorpus/ErrorSentenceTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-dev/src/test/java/org/languagetool/dev/errorcorpus/ErrorSentenceTest.java :: ErrorSentenceTest.testHasErrorCoveredByMatch
func TestErrorSentence_HasErrorCoveredByMatch(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
}

// Port of languagetool-dev/src/test/java/org/languagetool/dev/errorcorpus/ErrorSentenceTest.java :: ErrorSentenceTest.testHasErrorOverlappingWithMatch
func TestErrorSentence_HasErrorOverlappingWithMatch(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
}
