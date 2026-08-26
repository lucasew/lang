package patterns

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseUserInfo(t *testing.T) {
	c, err := ParseUserInfo("alice:s3cret")
	require.NoError(t, err)
	require.Equal(t, "alice", c.Username)
	require.Equal(t, "s3cret", c.Password)
	_, err = ParseUserInfo("bad")
	require.Error(t, err)
	c, err = ParseUserInfo("")
	require.NoError(t, err)
	require.Nil(t, c)
}

func TestGetPasswordAuthenticationFromURL(t *testing.T) {
	u, err := url.Parse("http://user:pass@example.com/path")
	require.NoError(t, err)
	c, err := GetPasswordAuthenticationFromURL(u)
	require.NoError(t, err)
	require.Equal(t, "user", c.Username)
	require.Equal(t, "pass", c.Password)
}
