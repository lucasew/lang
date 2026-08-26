package language

// Twin of BosnianSerbianConcurrencyTest — concurrent Analyze via languagetool package.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Port of BosnianSerbianConcurrencyTest (Java slow @Ignore spell race deferred)
func TestBosnianSerbianConcurrency_NoTests(t *testing.T) {
	languagetool.ConcurrencyAnalyzeSmoke(t, "sr-BA", "Тест.")
}
