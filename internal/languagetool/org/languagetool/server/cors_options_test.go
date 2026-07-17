package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServeHTTP_CORSOptions(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PublicAccess = true
	cfg.AllowOriginURL = "https://example.com"
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)

	req := httptest.NewRequest(http.MethodOptions, "/v2/check", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}
