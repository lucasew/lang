package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/ServerToolsTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-server/src/test/java/org/languagetool/server/ServerToolsTest.java :: ServerToolsTest.testCleanUserTextFromMessage
func TestServerTools_CleanUserTextFromMessage(t *testing.T) {
	// contains assertThat
}
