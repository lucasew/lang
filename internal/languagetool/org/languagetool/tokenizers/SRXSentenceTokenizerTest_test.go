package tokenizers

// Twin of languagetool-standalone/src/test/java/org/languagetool/tokenizers/SRXSentenceTokenizerTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-standalone/src/test/java/org/languagetool/tokenizers/SRXSentenceTokenizerTest.java :: SRXSentenceTokenizerTest.testOfficeFootnoteTokenize
func TestSRXSentenceTokenizer_OfficeFootnoteTokenize(t *testing.T) {
	t.Skip("unimplemented: SRXSentenceTokenizerTest.testOfficeFootnoteTokenize")
}

// Port of languagetool-standalone/src/test/java/org/languagetool/tokenizers/SRXSentenceTokenizerTest.java :: SRXSentenceTokenizerTest.testDotNetSentence
func TestSRXSentenceTokenizer_DotNetSentence(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}
