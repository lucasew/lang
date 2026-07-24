package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUUIDTypeHandler(t *testing.T) {
	u := UUIDBits{MostSignificant: 0x1234567890abcdef, LeastSignificant: 0xfedcba0987654321}
	b := UUIDBitsToBytes(u)
	require.Len(t, b, 16)
	back, err := BytesToUUIDBits(b)
	require.NoError(t, err)
	require.Equal(t, u, back)
	s := u.String()
	require.Contains(t, s, "-")
	parsed, err := ParseUUIDString(s)
	require.NoError(t, err)
	require.Equal(t, u, parsed)
}

func TestAPINewGroupAndLogging(t *testing.T) {
	g := NewAPINewGroup("team")
	require.Equal(t, "team", g.Name)
	require.True(t, g.Equal(APINewGroup{Name: "team"}))

	li := NewLoggingInterceptor()
	require.NoError(t, li.Intercept("select  *  from t", 1, func() error {
		time.Sleep(time.Millisecond)
		return nil
	}))
	require.Equal(t, 1, li.Len())
	require.Equal(t, "select * from t", li.Entries[0].SQL)

	app := NewInstrumentedAppender()
	app.Append("INFO", "org.lt", "", "")
	require.Equal(t, int64(1), app.Count("INFO", "org.lt"))
}

func TestHTTPTestToolsAndGRPC(t *testing.T) {
	require.Equal(t, DefaultPort, GetDefaultPort())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	}))
	defer srv.Close()
	body, err := CheckAtURL(srv.URL)
	require.NoError(t, err)
	require.Equal(t, "pong", body)
	body, err = CheckAtURLByPost(srv.URL, FormEncode(map[string]string{"a": "b"}), nil)
	require.NoError(t, err)
	require.Equal(t, "pong", body)

	g := NewGRPCServer()
	_, _, err = g.Analyze(ProcessingOptions{Language: "en"}, "hi")
	require.Error(t, err)
	g.InitPool(NewHTTPServerConfig())
	lang, n, err := g.Analyze(ProcessingOptions{Language: "en"}, "hello world")
	require.NoError(t, err)
	require.Equal(t, "en", lang)
	// Tokenization fidelity is separate; require analysis produces at least one non-boundary token.
	require.GreaterOrEqual(t, n, 1)
}
