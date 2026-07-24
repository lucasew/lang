package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/HTTPSServerTest.java
import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of HTTPSServerTest.runRequestLimitationTest
func TestHTTPSServer_RunRequestLimitationTest(t *testing.T) {
	cfg := NewHTTPSServerConfig("/tmp/ks.jks", "pw")
	cfg.RequestLimit = 1
	cfg.RequestLimitPeriodInSeconds = 60
	lim := NewRequestLimiter(1, 60)
	h := NewLanguageToolHttpHandler(cfg.HTTPServerConfig, nil, false, lim, nil, nil)
	_, err := h.HandlePath("/v2/languages", "9.9.9.9", nil)
	require.NoError(t, err)
	_, err = h.HandlePath("/v2/languages", "9.9.9.9", nil)
	require.Error(t, err)
	var tooMany *TooManyRequestsError
	require.ErrorAs(t, err, &tooMany)
}

// Port of HTTPSServerTest.runReferrerLimitationTest
func TestHTTPSServer_RunReferrerLimitationTest(t *testing.T) {
	cfg := NewHTTPSServerConfig("/tmp/ks.jks", "pw")
	cfg.SetBlockedReferrers([]string{"http://foo.org", "bar.org"})
	require.True(t, cfg.IsBlockedReferrer("http://foo.org/page"))
	require.True(t, cfg.IsBlockedReferrer("https://bar.org/x"))
	require.False(t, cfg.IsBlockedReferrer("https://ok.example/"))

	h := NewLanguageToolHttpHandler(cfg.HTTPServerConfig, nil, false, nil, nil, nil)
	r, err := h.HandlePathWithReferrer("/v2/languages", "127.0.0.1", "http://foo.org", nil)
	require.NoError(t, err)
	require.Equal(t, 403, r.Status)

	r, err = h.HandlePathWithReferrer("/v2/languages", "127.0.0.1", "https://ok.example/", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
}

// Port of HTTPSServerTest.testHTTPSServer — surface without live TLS bind
func TestHTTPSServer_HTTPSServer(t *testing.T) {
	cfg := NewHTTPSServerConfigPort(8443, false, "/path/ks.jks", "secret")
	s := NewHTTPSServer(cfg, false, "localhost", DefaultAllowedIPs)
	require.Equal(t, "https", s.Protocol())
	require.True(t, s.HasKeystore())
	require.Equal(t, 8443, s.TLSConfig.Port)
	// check still routes through handler
	h := NewLanguageToolHttpHandler(cfg.HTTPServerConfig, DefaultAllowedIPs, false, nil, nil, nil)
	q := url.Values{}
	r, err := h.HandlePath("/v2/info", "127.0.0.1", q)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
}
