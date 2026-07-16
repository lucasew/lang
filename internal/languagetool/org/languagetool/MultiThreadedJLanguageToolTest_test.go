package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/MultiThreadedJLanguageToolTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/MultiThreadedJLanguageToolTest.java :: MultiThreadedJLanguageToolTest.testCheck
func TestMultiThreadedJLanguageTool_Check(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-core/src/test/java/org/languagetool/MultiThreadedJLanguageToolTest.java :: MultiThreadedJLanguageToolTest.testShutdownException
func TestMultiThreadedJLanguageTool_ShutdownException(t *testing.T) {
	t.Skip("unimplemented: MultiThreadedJLanguageToolTest.testShutdownException")
}

// Port of languagetool-core/src/test/java/org/languagetool/MultiThreadedJLanguageToolTest.java :: MultiThreadedJLanguageToolTest.testTextAnalysis
func TestMultiThreadedJLanguageTool_TextAnalysis(t *testing.T) {
	// contains assertThat
}

// Port of languagetool-core/src/test/java/org/languagetool/MultiThreadedJLanguageToolTest.java :: MultiThreadedJLanguageToolTest.testConfigurableThreadPoolSize
func TestMultiThreadedJLanguageTool_ConfigurableThreadPoolSize(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-core/src/test/java/org/languagetool/MultiThreadedJLanguageToolTest.java :: MultiThreadedJLanguageToolTest.testTwoRulesOnly
func TestMultiThreadedJLanguageTool_TwoRulesOnly(t *testing.T) {
	// contains assertThat
}
