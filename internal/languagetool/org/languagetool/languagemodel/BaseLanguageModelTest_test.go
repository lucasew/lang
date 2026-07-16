package languagemodel

// Twin of languagetool-core/src/test/java/org/languagetool/languagemodel/BaseLanguageModelTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/languagemodel/BaseLanguageModelTest.java :: BaseLanguageModelTest.testPseudoProbability
func TestBaseLanguageModel_PseudoProbability(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertThat
}

// Port of languagetool-core/src/test/java/org/languagetool/languagemodel/BaseLanguageModelTest.java :: BaseLanguageModelTest.testPseudoProbabilityFail1
func TestBaseLanguageModel_PseudoProbabilityFail1(t *testing.T) {
	tools.Unimplemented("BaseLanguageModelTest.testPseudoProbabilityFail1")
}
