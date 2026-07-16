package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java :: TextCheckerTest.testJSONP
func TestTextChecker_JSONP(t *testing.T) {
	// contains assertTrue
}

// Port of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java :: TextCheckerTest.testMaxTextLength
func TestTextChecker_MaxTextLength(t *testing.T) {
	tools.Unimplemented("TextCheckerTest.testMaxTextLength")
}

// Port of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java :: TextCheckerTest.testInvalidAltLanguages
func TestTextChecker_InvalidAltLanguages(t *testing.T) {
	tools.Unimplemented("TextCheckerTest.testInvalidAltLanguages")
}

// Port of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java :: TextCheckerTest.testDetectLanguageOfString
func TestTextChecker_DetectLanguageOfString(t *testing.T) {
	t.Skip("Java @Ignore")
	// contains assertTrue
	// contains assertThat
}

// Port of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java :: TextCheckerTest.testInvalidPreferredVariant
func TestTextChecker_InvalidPreferredVariant(t *testing.T) {
	tools.Unimplemented("TextCheckerTest.testInvalidPreferredVariant")
}

// Port of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java :: TextCheckerTest.testInvalidPreferredVariant2
func TestTextChecker_InvalidPreferredVariant2(t *testing.T) {
	tools.Unimplemented("TextCheckerTest.testInvalidPreferredVariant2")
}
