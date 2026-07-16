package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/RemoteSynthesizerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-server/src/test/java/org/languagetool/server/RemoteSynthesizerTest.java :: RemoteSynthesizerTest.testSynthesis
func TestRemoteSynthesizer_Synthesis(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertNull
}
