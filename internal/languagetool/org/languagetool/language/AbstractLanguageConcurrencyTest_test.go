package language

// Twin of languagetool-core/src/test/java/org/languagetool/language/AbstractLanguageConcurrencyTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/language/AbstractLanguageConcurrencyTest.java :: AbstractLanguageConcurrencyTest.testSpellCheckerFailure
func TestAbstractLanguageConcurrency_SpellCheckerFailure(t *testing.T) {
	t.Skip("Java @Ignore")
	// contains assertEquals — full values in Java twin source
}
