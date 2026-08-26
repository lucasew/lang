package language

// Twin of CroatianSerbianConcurrencyTest — concurrent Analyze via languagetool package.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Port of CroatianSerbianConcurrencyTest (Java slow @Ignore spell race deferred)
func TestCroatianSerbianConcurrency_NoTests(t *testing.T) {
	languagetool.ConcurrencyAnalyzeSmoke(t, "sr-HR", "Тест.")
}
