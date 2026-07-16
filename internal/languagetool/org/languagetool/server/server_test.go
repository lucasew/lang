package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestLimiter(t *testing.T) {
	l := NewRequestLimiter(2, 60)
	require.True(t, l.Allow("1.1.1.1"))
	require.True(t, l.Allow("1.1.1.1"))
	require.False(t, l.Allow("1.1.1.1"))
	require.True(t, l.Allow("2.2.2.2"))
}

func TestServerTools(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "9.9.9.9, 8.8.8.8")
	require.Equal(t, "10.0.0.1", GetHTTPRequestIP(r, false))
	require.Equal(t, "9.9.9.9", GetHTTPRequestIP(r, true))
	require.Contains(t, CleanUserQuery("a\nb", 10), "a")
	require.Equal(t, "LanguageTool-Go", NewSoftwareInfo("").Name)
}
