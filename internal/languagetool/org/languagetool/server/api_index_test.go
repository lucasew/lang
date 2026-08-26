package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPHandler_APIIndex(t *testing.T) {
	cfg := NewHTTPServerConfig()
	h := NewLanguageToolHttpHandler(cfg, map[string]struct{}{"127.0.0.1": {}}, false, nil, nil, nil)
	for _, path := range []string{"/", "/v2", "/v2/"} {
		r, err := h.HandlePath(path, "127.0.0.1", nil)
		require.NoError(t, err, path)
		require.Equal(t, 200, r.Status)
		require.Contains(t, r.Body, "/v2/check")
		require.Contains(t, r.Body, "endpoints")
	}
}
