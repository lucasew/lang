package language

// Twin of languagetool-standalone/src/test/java/org/languagetool/language/LanguageIdentifierServiceTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-standalone/src/test/java/org/languagetool/language/LanguageIdentifierServiceTest.java :: LanguageIdentifierServiceTest.testFactory
func TestLanguageIdentifierService_Factory(t *testing.T) {
	// contains assertTrue
}

// Port of languagetool-standalone/src/test/java/org/languagetool/language/LanguageIdentifierServiceTest.java :: LanguageIdentifierServiceTest.testFactoryWithoutReset
func TestLanguageIdentifierService_FactoryWithoutReset(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}
