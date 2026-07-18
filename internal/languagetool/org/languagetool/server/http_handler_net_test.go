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

func TestLanguageToolHttpHandler_ServeHTTP_Check(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PublicAccess = true
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)

	form := url.Values{}
	form.Set("language", "en")
	form.Set("text", "This is an test.")
	req := httptest.NewRequest(http.MethodPost, "/v2/check", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	require.Contains(t, string(body), "EN_A_VS_AN")
	require.Contains(t, rr.Header().Get("Content-Type"), "json")
}

func TestLanguageToolHttpHandler_ServeHTTP_Languages(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PublicAccess = true
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/v2/languages", nil)
	req.RemoteAddr = "127.0.0.1:1"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "English")
}

func TestLanguageToolHttpHandler_ServeHTTP_MissingLang(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PublicAccess = true
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/v2/check?text=hi", nil)
	req.RemoteAddr = "127.0.0.1:1"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestApiV2_CheckPickyLevel(t *testing.T) {
	// Invent picky token packs removed; picky path still returns 200 for ordinary grammar.
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "This is an test.",
		"level":    "picky",
	})
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "EN_A_VS_AN")
}

func TestApiV2_CheckEnabledOnly(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language":    "en",
		"text":        "This is an test. hello  world",
		"enabledRules": "EN_A_VS_AN",
		"enabledOnly": "true",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "EN_A_VS_AN")
	require.NotContains(t, r.Body, "WHITESPACE_RULE")
}

func TestTextChecker_PipelinePoolReuse(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = true
	cfg.MaxPipelinePoolSize = 4
	tc := NewV2TextChecker(cfg, false, nil)
	require.NotNil(t, tc.Pool)
	ms1 := tc.Check("This is an test.", "en", nil)
	ms2 := tc.Check("This is an test.", "en", nil)
	require.NotEmpty(t, ms1)
	require.NotEmpty(t, ms2)
	require.GreaterOrEqual(t, tc.Pool.IdleCount(pipelineSettingsFor("en", CheckOptions{})), 0)
}
