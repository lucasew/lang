package languagetool

// Twin of languagetool-standalone/src/test/java/org/languagetool/ShortDescriptionProviderTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-standalone/src/test/java/org/languagetool/ShortDescriptionProviderTest.java :: ShortDescriptionProviderTest.testGetShortDescription
func TestShortDescriptionProvider_GetShortDescription(t *testing.T) {
	// contains assertNull
	// contains assertNotNull
}

// Port of languagetool-standalone/src/test/java/org/languagetool/ShortDescriptionProviderTest.java :: ShortDescriptionProviderTest.testDescriptionLength
func TestShortDescriptionProvider_DescriptionLength(t *testing.T) {
	t.Skip("unimplemented: ShortDescriptionProviderTest.testDescriptionLength")
}
