package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/TrustedSourcesTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-server/src/test/java/org/languagetool/server/TrustedSourcesTest.java :: TrustedSourcesTest.runUntrustedReferrerTest
func TestTrustedSources_RunUntrustedReferrerTest(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
}
