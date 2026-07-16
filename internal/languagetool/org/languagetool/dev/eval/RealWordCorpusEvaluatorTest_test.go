package eval

// Twin of languagetool-dev/src/test/java/org/languagetool/dev/eval/RealWordCorpusEvaluatorTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-dev/src/test/java/org/languagetool/dev/eval/RealWordCorpusEvaluatorTest.java :: RealWordCorpusEvaluatorTest.testCheck
func TestRealWordCorpusEvaluator_Check(t *testing.T) {
	t.Skip("Java @Ignore")
	// contains assertThat
}
