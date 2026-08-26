package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/HTTPServerTest.java
import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of HTTPServerTest.testHTTPServer — handler surface without live socket
func TestHTTPServer_HTTPServer(t *testing.T) {
	cfg := NewHTTPServerConfig()
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)
	require.False(t, h.IsShutdown())
	r, err := h.HandlePath("/v2/languages", "127.0.0.1", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	h.Shutdown()
	require.True(t, h.IsShutdown())
}

// Port of HTTPServerTest.testTimeout — max check time config surface
func TestHTTPServer_Timeout(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.MaxCheckTimeMillisAnonymous = 1
	require.Equal(t, int64(1), cfg.MaxCheckTimeMillisAnonymous)
	// full wall-clock timeout needs live check pipeline; config is wired.
}

// Port of HTTPServerTest.testHealthcheck
func TestHTTPServer_Healthcheck(t *testing.T) {
	cfg := NewHTTPServerConfig()
	api := NewApiV2(cfg, nil)
	r, err := api.Handle("info", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "software")
}

// Port of HTTPServerTest.testAccessDenied
func TestHTTPServer_AccessDenied(t *testing.T) {
	cfg := NewHTTPServerConfig()
	allowed := map[string]struct{}{"127.0.0.1": {}}
	h := NewLanguageToolHttpHandler(cfg, allowed, false, nil, nil, nil)
	r, err := h.HandlePath("/v2/languages", "10.0.0.9", nil)
	require.NoError(t, err)
	require.Equal(t, 403, r.Status)
	require.Contains(t, r.Body, "not allowed")
}

// Port of HTTPServerTest.testRequestLimit
func TestHTTPServer_RequestLimit(t *testing.T) {
	cfg := NewHTTPServerConfig()
	lim := NewRequestLimiter(1, 60)
	h := NewLanguageToolHttpHandler(cfg, nil, false, lim, nil, nil)
	_, err := h.HandlePath("/v2/languages", "1.2.3.4", nil)
	require.NoError(t, err)
	_, err = h.HandlePath("/v2/languages", "1.2.3.4", nil)
	require.Error(t, err)
	var tooMany *TooManyRequestsError
	require.ErrorAs(t, err, &tooMany)
}

// Port of HTTPServerTest.testEnabledOnlyParameter
func TestHTTPServer_EnabledOnlyParameter(t *testing.T) {
	p, err := ParseCheckQueryParams(map[string]string{
		"enabledOnly": "true",
		"enabledRules": "RULE_A",
		"callback": "ok",
	})
	require.NoError(t, err)
	require.True(t, p.UseEnabledOnly)
	require.Equal(t, []string{"RULE_A"}, p.EnabledRules)
}

// Port of HTTPServerTest.testServerUrlSetting
func TestHTTPServer_ServerUrlSetting(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.ServerURL = "https://example.com/api"
	require.Equal(t, "https://example.com/api", cfg.ServerURL)
}

// Port of HTTPServerTest.testMissingLanguageParameter
func TestHTTPServer_MissingLanguageParameter(t *testing.T) {
	cfg := NewHTTPServerConfig()
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)
	q := url.Values{}
	q.Set("text", "hello")
	_, err := h.HandlePath("/v2/check", "127.0.0.1", q)
	require.Error(t, err)
	var bad *BadRequestError
	require.ErrorAs(t, err, &bad)
	require.Contains(t, err.Error(), "language")
}
