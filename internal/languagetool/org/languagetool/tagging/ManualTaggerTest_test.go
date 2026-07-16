package tagging

// Twin of languagetool-standalone/src/test/java/org/languagetool/tagging/ManualTaggerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-standalone/src/test/java/org/languagetool/tagging/ManualTaggerTest.java :: ManualTaggerTest.testTag
func TestManualTagger_Tag(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertThat
}
