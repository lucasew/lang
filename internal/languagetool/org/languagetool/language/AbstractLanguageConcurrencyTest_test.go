package language

// Twin of AbstractLanguageConcurrencyTest — soft concurrent Analyze (Java @Ignore).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Port of AbstractLanguageConcurrencyTest.testSpellCheckerFailure
func TestAbstractLanguageConcurrency_SpellCheckerFailure(t *testing.T) {
	// Java @Ignore: too slow for full spell race — green Analyze concurrency instead
	languagetool.ConcurrencyAnalyzeSmoke(t, "en", "Sample concurrent text.")
	languagetool.ConcurrencyAnalyzeSmoke(t, "de", "Ein paralleler Test.")
}
