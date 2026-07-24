package remote

import (
	"net/http/httptest"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/server"
	"github.com/stretchr/testify/require"
)

// End-to-end: RemoteLanguageTool client against pure-Go HTTP handler.
func TestRemoteLanguageTool_LocalServer(t *testing.T) {
	cfg := server.NewHTTPServerConfig()
	cfg.PublicAccess = true
	h := server.NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	// httptest URL has no trailing slash; RemoteLanguageTool requires no trailing slash
	lt := NewRemoteLanguageTool(srv.URL)
	res, err := lt.Check("This is an test.", "en")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "en", res.GetLanguageCode())
	require.NotEmpty(t, res.GetMatches())
	found := false
	for _, m := range res.GetMatches() {
		if m.GetRuleID() == "EN_A_VS_AN" {
			found = true
		}
	}
	require.True(t, found, "%+v", res.GetMatches())
}
