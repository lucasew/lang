package language

// Twin of SerbianConcurrencyTest — concurrent Analyze via languagetool package.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Port of SerbianConcurrencyTest (Java slow @Ignore spell race deferred)
func TestSerbianConcurrency_NoTests(t *testing.T) {
	languagetool.ConcurrencyAnalyzeSmoke(t, "sr", "Тест реченица.")
}
