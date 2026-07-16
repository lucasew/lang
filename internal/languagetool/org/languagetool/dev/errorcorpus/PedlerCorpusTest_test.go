package errorcorpus

// Twin of languagetool-dev/src/test/java/org/languagetool/dev/errorcorpus/PedlerCorpusTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-dev/src/test/java/org/languagetool/dev/errorcorpus/PedlerCorpusTest.java :: PedlerCorpusTest.testCorpusAccess
func TestPedlerCorpus_CorpusAccess(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
	// contains assertThat
}
