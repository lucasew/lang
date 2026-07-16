package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/TrustedSourcesTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of TrustedSourcesTest.runUntrustedReferrerTest
func TestTrustedSources_RunUntrustedReferrerTest(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.SetBlockedReferrers([]string{"evil.example", "spam.org"})
	require.True(t, cfg.IsBlockedReferrer("https://evil.example/ref"))
	require.False(t, cfg.IsBlockedReferrer("https://trusted.example/"))
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)
	r, err := h.HandlePathWithReferrer("/v2/languages", "1.1.1.1", "https://spam.org/x", nil)
	require.NoError(t, err)
	require.Equal(t, 403, r.Status)
}
