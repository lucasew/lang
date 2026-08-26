package language

// Twin of MontenegrinSerbianConcurrencyTest — concurrent Analyze via languagetool package.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Port of MontenegrinSerbianConcurrencyTest (Java slow @Ignore spell race deferred)
func TestMontenegrinSerbianConcurrency_NoTests(t *testing.T) {
	languagetool.ConcurrencyAnalyzeSmoke(t, "sr-ME", "Тест.")
}
