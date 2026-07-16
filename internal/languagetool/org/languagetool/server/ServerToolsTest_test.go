package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/ServerToolsTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of languagetool-server/src/test/java/org/languagetool/server/ServerToolsTest.java :: ServerToolsTest.testCleanUserTextFromMessage
func TestServerTools_CleanUserTextFromMessage(t *testing.T) {
	loggingOn := map[string]string{}
	loggingOff := map[string]string{"inputLogging": "no"}
	require.Equal(t, "my test", CleanUserTextFromMessage("my test", loggingOn))
	require.Equal(t, "my test", CleanUserTextFromMessage("my test", loggingOff))
	require.Equal(t, "<sentcontent>my test</sentcontent>", CleanUserTextFromMessage("<sentcontent>my test</sentcontent>", loggingOn))
	require.Equal(t, "<< content removed >>", CleanUserTextFromMessage("<sentcontent>my test</sentcontent>", loggingOff))
	require.Equal(t, "<< content removed >>", CleanUserTextFromMessage("<sentcontent>my\ntest</sentcontent>", loggingOff))
}
