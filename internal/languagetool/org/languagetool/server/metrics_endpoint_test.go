package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiV2_Metrics(t *testing.T) {
	api := NewApiV2(nil, nil)
	// produce some check traffic
	_, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "This is an test.",
	})
	require.NoError(t, err)

	r, err := api.Handle("metrics", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	var snap MetricsSnapshot
	require.NoError(t, json.Unmarshal([]byte(r.Body), &snap))
	require.GreaterOrEqual(t, snap.Checks, int64(1))
	require.GreaterOrEqual(t, snap.HTTPRequests, int64(1))
	require.GreaterOrEqual(t, snap.Matches, int64(1))
	require.NotEmpty(t, snap.ChecksByLanguage)
}

func TestHTTP_E2E_Metrics(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PublicAccess = true
	h := NewLanguageToolHttpHandler(cfg, nil, false, nil, nil, nil)

	form := url.Values{}
	form.Set("language", "en")
	form.Set("text", "ok")
	req := httptest.NewRequest(http.MethodPost, "/v2/check", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/v2/metrics", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "checks")
	require.Contains(t, w.Body.String(), "httpRequests")
}
