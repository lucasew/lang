package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/ResourceBundleToolsTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/ResourceBundleToolsTest.java :: ResourceBundleToolsTest.testGetMessageBundle
func TestResourceBundleTools_GetMessageBundle(t *testing.T) {
	// contains assertNotNull
}
