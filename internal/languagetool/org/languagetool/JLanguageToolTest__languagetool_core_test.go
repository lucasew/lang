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
	lt.AddChecker(SimpleWordRepeatChecker(""))
	require.Empty(t, lt.Check("This is a test."))
	require.NotEmpty(t, lt.Check("This is is a test."))
}

// Port of JLanguageToolTest.testCheckStringWithCallbackReturnsTrue
func TestJLanguageTool_languagetool_core_CheckStringWithCallbackReturnsTrue(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddChecker(SimpleWordRepeatChecker(""))
	lt.Cancelled = func() bool { return false }
	require.NotEmpty(t, lt.Check("is is wrong"))
}

// Port of JLanguageToolTest.testCheckStringWithCallbackReturnsFalse
func TestJLanguageTool_languagetool_core_CheckStringWithCallbackReturnsFalse(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddChecker(SimpleWordRepeatChecker(""))
	lt.Cancelled = func() bool { return true }
	require.Empty(t, lt.Check("is is wrong"))
}
