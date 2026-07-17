package server

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPServerAndHandler(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.MaxTextLengthAnonymous = 1000
	cfg.RequestLimit = 100
	cfg.RequestLimitPeriodInSeconds = 60
	srv := NewHTTPServerWithConfig(cfg)
	require.Equal(t, "http", srv.Protocol())
	require.True(t, srv.AllowIP("127.0.0.1"))
	require.False(t, srv.AllowIP("8.8.8.8"))

	h := srv.Handler
	require.False(t, h.IsShutdown())

	q := url.Values{}
	q.Set("language", "en")
	q.Set("text", "Hello")
	r, err := h.HandlePath("/v2/check", "127.0.0.1", q)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "matches")

	// live engine via handler
	q2 := url.Values{}
	q2.Set("language", "en")
	q2.Set("text", "This is an test.")
	r2, err := h.HandlePath("/v2/check", "127.0.0.1", q2)
	require.NoError(t, err)
	require.Equal(t, 200, r2.Status)
	require.Contains(t, r2.Body, "EN_A_VS_AN")

	r, err = h.HandlePath("/v2/languages", "127.0.0.1", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)

	_, err = h.HandlePath("/v2/check", "8.8.8.8", q)
	// IP not allowed returns 403 result without error
	require.NoError(t, err)

	h.Shutdown()
	_, err = h.HandlePath("/v2/check", "127.0.0.1", q)
	require.Error(t, err)
}

func TestRemoteSynthesizerAndDBLog(t *testing.T) {
	rs := NewRemoteSynthesizer(func(lang, lemma, postag string, re bool) ([]string, error) {
		return []string{"a", "b", "a"}, nil
	})
	forms, err := rs.SynthesizeForms("en", "go", "VB", false)
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, forms)

	uid := int64(3)
	ping := NewDatabasePingLogEntry(nil, &uid)
	require.Equal(t, "org.languagetool.server.LogMapper.pings", ping.GetMappingIdentifier())
	require.Equal(t, int64(3), ping.GetMapping()["user_id"])
	require.Nil(t, ping.Followup())

	check := NewDatabaseCheckLogEntry(&uid, 12, "en", 2)
	require.Equal(t, 12, check.GetMapping()["text_size"])
	require.Contains(t, check.GetMappingIdentifier(), "checks")
}
