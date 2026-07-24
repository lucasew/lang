package remote

// Twin of languagetool-http-client/src/test/java/org/languagetool/remote/RemoteLanguageToolIntegrationTest.java
import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of RemoteLanguageToolIntegrationTest.testClient (offline: parse + params; live HTTP deferred)
func TestRemoteLanguageToolIntegration_Client(t *testing.T) {
	// Offline surface already covered in remote_test.go; assert constructor + Check wiring.
	body := `{"software":{"name":"LanguageTool","version":"6.0"},"language":{"name":"English","code":"en"},"matches":[{"message":"use a","offset":2,"length":1,"context":{"text":"A a","offset":2,"length":1},"replacements":[{"value":"an"}],"rule":{"id":"EN_A_VS_AN","description":"a vs an","category":{"id":"G","name":"Grammar"}}}]}`
	lt := NewRemoteLanguageTool("http://127.0.0.1:8081")
	lt.Client = stubClient{body: body}
	res, err := lt.Check("A a", "en")
	require.NoError(t, err)
	require.Equal(t, "en", res.GetLanguageCode())
	require.Equal(t, "LanguageTool", res.GetRemoteServer().GetSoftware())
	require.Len(t, res.GetMatches(), 1)
	require.Equal(t, "EN_A_VS_AN", res.GetMatches()[0].GetRuleID())

	cfg := NewCheckConfigurationBuilder("en").DisabledRuleIDs("EN_A_VS_AN").Build()
	params := GetURLParams("text", cfg, nil)
	require.Equal(t, "EN_A_VS_AN", params.Get("disabledRules"))
}

// Port of RemoteLanguageToolIntegrationTest.testClientWithHTTPS
func TestRemoteLanguageToolIntegration_ClientWithHTTPS(t *testing.T) {
	// Constructor accepts https; no live TLS server in unit tests.
	lt := NewRemoteLanguageTool("https://127.0.0.1:8443")
	require.Equal(t, "https://127.0.0.1:8443", lt.ServerBaseURL)
	lt.Client = stubClient{body: `{"software":{"name":"LT","version":"1"},"language":{"name":"English","code":"en"},"matches":[]}`}
	res, err := lt.Check("ok", "en")
	require.NoError(t, err)
	require.Empty(t, res.GetMatches())
}

// Port of RemoteLanguageToolIntegrationTest.testInvalidServer
func TestRemoteLanguageToolIntegration_InvalidServer(t *testing.T) {
	lt := NewRemoteLanguageTool("http://does-not-exist.invalid")
	lt.Client = errClient{err: io.EOF}
	_, err := lt.Check("foo", "en")
	require.Error(t, err)
}

// Port of RemoteLanguageToolIntegrationTest.testWrongProtocol
func TestRemoteLanguageToolIntegration_WrongProtocol(t *testing.T) {
	// HTTPS client against unreachable/TLS-mismatched endpoint surfaces as Check error.
	lt := NewRemoteLanguageTool("https://127.0.0.1:1")
	lt.Client = errClient{err: io.ErrUnexpectedEOF}
	_, err := lt.Check("foo", "en")
	require.Error(t, err)
}

// Port of RemoteLanguageToolIntegrationTest.testInvalidProtocol
func TestRemoteLanguageToolIntegration_InvalidProtocol(t *testing.T) {
	require.Panics(t, func() {
		NewRemoteLanguageTool("ftp://127.0.0.1:8081")
	})
}

// Port of RemoteLanguageToolIntegrationTest.testProtocolTypo
func TestRemoteLanguageToolIntegration_ProtocolTypo(t *testing.T) {
	require.Panics(t, func() {
		NewRemoteLanguageTool("htp://127.0.0.1:8081")
	})
}

type errClient struct{ err error }

func (e errClient) Do(req *http.Request) (*http.Response, error) {
	return nil, e.err
}

// ensure stubClient from remote_test.go is available (same package)
var _ = strings.TrimSpace
