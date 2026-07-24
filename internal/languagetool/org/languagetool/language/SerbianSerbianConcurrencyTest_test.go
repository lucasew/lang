package language

// Twin of SerbianSerbianConcurrencyTest — concurrent Analyze via languagetool package.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Port of SerbianSerbianConcurrencyTest (Java slow @Ignore spell race deferred)
func TestSerbianSerbianConcurrency_NoTests(t *testing.T) {
	languagetool.ConcurrencyAnalyzeSmoke(t, "sr-RS", "Тест.")
}
