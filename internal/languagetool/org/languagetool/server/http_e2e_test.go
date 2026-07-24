package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTP_E2E_CheckAndLanguages(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PublicAccess = true
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)

	// index
	req := httptest.NewRequest(http.MethodGet, "/v2", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "/v2/check")
	require.Equal(t, "LanguageTool-Go", w.Header().Get("X-LanguageTool-Software"))
	require.Equal(t, "1", w.Header().Get("X-LanguageTool-API-Version"))
	require.NotEmpty(t, w.Header().Get("X-Request-ID"))
	require.NotEmpty(t, w.Header().Get("X-LanguageTool-Time-ms"))

	// languages
	req = httptest.NewRequest(http.MethodGet, "/v2/languages", nil)
	req.Header.Set("X-Request-ID", "client-req-1")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "en-US")
	require.Equal(t, "client-req-1", w.Header().Get("X-Request-ID"))

	// check via POST form
	form := url.Values{}
	form.Set("language", "en")
	form.Set("text", "This is an test.")
	req = httptest.NewRequest(http.MethodPost, "/v2/check", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	body, _ := io.ReadAll(w.Body)
	require.Contains(t, string(body), "EN_A_VS_AN")
	// Java AvsAnRule: setLocQualityIssueType(ITSIssueType.Misspelling)
	require.Contains(t, string(body), `"typeName":"misspelling"`)
}
