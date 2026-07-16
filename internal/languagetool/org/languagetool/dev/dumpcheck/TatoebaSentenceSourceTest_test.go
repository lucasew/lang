package dumpcheck

// Twin of languagetool-wikipedia/src/test/java/org/languagetool/dev/dumpcheck/TatoebaSentenceSourceTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-wikipedia/src/test/java/org/languagetool/dev/dumpcheck/TatoebaSentenceSourceTest.java :: TatoebaSentenceSourceTest.testTatoebaSource
func TestTatoebaSentenceSource_TatoebaSource(t *testing.T) {
	t.Skip("Java @Ignore")
	// contains assertTrue
	// contains assertFalse
	// contains assertThat
}

// Port of languagetool-wikipedia/src/test/java/org/languagetool/dev/dumpcheck/TatoebaSentenceSourceTest.java :: TatoebaSentenceSourceTest.testTatoebaSourceInvalidInput
func TestTatoebaSentenceSource_TatoebaSourceInvalidInput(t *testing.T) {
	tools.Unimplemented("TatoebaSentenceSourceTest.testTatoebaSourceInvalidInput")
}
