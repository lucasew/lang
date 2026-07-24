package remote

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPTools(t *testing.T) {
	require.Equal(t, "http://x/v2/check", JoinURL("http://x/", "/v2/check"))
	require.Equal(t, "http://x/v2", JoinURL("http://x", "v2"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()
	h := NewHTTPTools(0)
	body, err := h.GetString(srv.URL)
	require.NoError(t, err)
	require.Equal(t, "ok", body)
	body, code, err := h.PostForm(srv.URL, map[string]string{"a": "b"})
	require.NoError(t, err)
	require.Equal(t, 200, code)
	require.Equal(t, "ok", body)
}
