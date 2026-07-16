package language

// Twin of languagetool-core/src/test/java/org/languagetool/language/LanguageBuilderTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/language/LanguageBuilderTest.java :: LanguageBuilderTest.testMakeAdditionalLanguage
func TestLanguageBuilder_MakeAdditionalLanguage(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertTrue
}

// Port of languagetool-core/src/test/java/org/languagetool/language/LanguageBuilderTest.java :: LanguageBuilderTest.testIllegalFileName
func TestLanguageBuilder_IllegalFileName(t *testing.T) {
	tools.Unimplemented("LanguageBuilderTest.testIllegalFileName")
}
