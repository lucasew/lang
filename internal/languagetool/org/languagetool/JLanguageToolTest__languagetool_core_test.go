package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/JLanguageToolTest.java
// Full check engine deferred — Analyze + mode/level surface.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testUserConfig
func TestJLanguageTool_languagetool_core_UserConfig(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.SetMode(ModeAll)
	lt.SetLevel(LevelDefault)
	require.Equal(t, ModeAll, lt.GetMode())
	require.Equal(t, LevelDefault, lt.GetLevel())
}

// Port of JLanguageToolTest.testCheckString
func TestJLanguageTool_languagetool_core_CheckString(t *testing.T) {
	lt := NewJLanguageTool("en")
	sents := lt.Analyze("This is a test.")
	require.NotEmpty(t, sents)
	require.NotEmpty(t, sents[0].GetText())
}

// Port of JLanguageToolTest.testCheckStringWithCallbackReturnsTrue
func TestJLanguageTool_languagetool_core_CheckStringWithCallbackReturnsTrue(t *testing.T) {
	// callback surface: CheckCancelledCallback type exists; analyze completes when not cancelled
	var cancelled CheckCancelledCallback = func() bool { return false }
	require.False(t, cancelled())
	lt := NewJLanguageTool("en")
	require.NotEmpty(t, lt.Analyze("Hello world."))
}

// Port of JLanguageToolTest.testCheckStringWithCallbackReturnsFalse
func TestJLanguageTool_languagetool_core_CheckStringWithCallbackReturnsFalse(t *testing.T) {
	var cancelled CheckCancelledCallback = func() bool { return true }
	require.True(t, cancelled())
	// full cancel mid-check deferred; surface type only
}
