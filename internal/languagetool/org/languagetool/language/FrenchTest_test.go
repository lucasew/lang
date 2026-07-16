package language

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/language/FrenchTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/fr/src/test/java/org/languagetool/language/FrenchTest.java :: FrenchTest.testSentenceTokenizer
func TestFrench_SentenceTokenizer(t *testing.T) {
	// contains assertThat
}

// Port of languagetool-language-modules/fr/src/test/java/org/languagetool/language/FrenchTest.java :: FrenchTest.testAdvancedTypography
func TestFrench_AdvancedTypography(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-language-modules/fr/src/test/java/org/languagetool/language/FrenchTest.java :: FrenchTest.testRules
func TestFrench_Rules(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}
