package remote

// Twin of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java :: RemoteLanguageToolIntegrationTest.testClient
func TestRemoteLanguageToolIntegration_Client(t *testing.T) {
	// contains assertTrue
	// contains assertThat
	// contains assertNotNull
}

// Port of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java :: RemoteLanguageToolIntegrationTest.testClientWithHTTPS
func TestRemoteLanguageToolIntegration_ClientWithHTTPS(t *testing.T) {
	// contains assertThat
}

// Port of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java :: RemoteLanguageToolIntegrationTest.testInvalidServer
func TestRemoteLanguageToolIntegration_InvalidServer(t *testing.T) {
	tools.Unimplemented("RemoteLanguageToolIntegrationTest.testInvalidServer")
}

// Port of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java :: RemoteLanguageToolIntegrationTest.testWrongProtocol
func TestRemoteLanguageToolIntegration_WrongProtocol(t *testing.T) {
	tools.Unimplemented("RemoteLanguageToolIntegrationTest.testWrongProtocol")
}

// Port of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java :: RemoteLanguageToolIntegrationTest.testInvalidProtocol
func TestRemoteLanguageToolIntegration_InvalidProtocol(t *testing.T) {
	tools.Unimplemented("RemoteLanguageToolIntegrationTest.testInvalidProtocol")
}

// Port of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java :: RemoteLanguageToolIntegrationTest.testProtocolTypo
func TestRemoteLanguageToolIntegration_ProtocolTypo(t *testing.T) {
	tools.Unimplemented("RemoteLanguageToolIntegrationTest.testProtocolTypo")
}
