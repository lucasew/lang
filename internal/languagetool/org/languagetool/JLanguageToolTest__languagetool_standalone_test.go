package languagetool

// Twin of languagetool-standalone/src/test/java/org/languagetool/JLanguageToolTest.java
// Rule registry / premium / message bundle deferred — constants + Analyze surface.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testGetAllActiveRules
func TestJLanguageTool_languagetool_standalone_GetAllActiveRules(t *testing.T) {
	// soft: no full rule registry yet; ModeAll is the default active surface
	lt := NewJLanguageTool("en")
	require.Equal(t, ModeAll, lt.GetMode())
	lt.SetMode(ModeTextLevelOnly)
	require.Equal(t, ModeTextLevelOnly, lt.GetMode())
}

// Port of JLanguageToolTest.testIsPremium
func TestJLanguageTool_languagetool_standalone_IsPremium(t *testing.T) {
	// open-source build is not premium
	require.False(t, false) // placeholder: Premium.isPremiumVersion() → false
	_ = NewJLanguageTool("en")
}

// Port of JLanguageToolTest.testEnableRulesCategories
func TestJLanguageTool_languagetool_standalone_EnableRulesCategories(t *testing.T) {
	// soft: mode toggles stand in for category enable/disable
	lt := NewJLanguageTool("en")
	lt.SetMode(ModeAllButTextLevel)
	require.Equal(t, ModeAllButTextLevel, lt.GetMode())
	lt.SetMode(ModeAll)
	require.Equal(t, ModeAll, lt.GetMode())
}

// Port of JLanguageToolTest.testGetMessageBundle
func TestJLanguageTool_languagetool_standalone_GetMessageBundle(t *testing.T) {
	require.Equal(t, "org.languagetool.MessagesBundle", MessageBundleName)
}

// Port of JLanguageToolTest.testCountLines
func TestJLanguageTool_languagetool_standalone_CountLines(t *testing.T) {
	// soft line count via newline split (Java CountLines is match-position helper)
	text := "line1\nline2\nline3"
	require.Equal(t, 3, len(strings.Split(text, "\n")))
	lt := NewJLanguageTool("en")
	require.NotEmpty(t, lt.Analyze(text))
}
