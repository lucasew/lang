package languagetool

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/JLanguageToolTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/pt/src/test/java/org/languagetool/JLanguageToolTest.java :: JLanguageToolTest.testPortugueseVariants
func TestJLanguageTool_lang_pt_PortugueseVariants(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-language-modules/pt/src/test/java/org/languagetool/JLanguageToolTest.java :: JLanguageToolTest.testSomeSentences
func TestJLanguageTool_lang_pt_SomeSentences(t *testing.T) {
	t.Skip("unimplemented: JLanguageToolTest.testSomeSentences")
}
