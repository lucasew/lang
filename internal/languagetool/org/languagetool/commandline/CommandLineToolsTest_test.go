package commandline

// Twin of languagetool-commandline/src/test/java/org/languagetool/commandline/CommandLineToolsTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-commandline/src/test/java/org/languagetool/commandline/CommandLineToolsTest.java :: CommandLineToolsTest.testCheck
func TestCommandLineTools_Check(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertTrue
}
