package server

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseBasicAuthentication(t *testing.T) {
	enc := base64.StdEncoding.EncodeToString([]byte("alice:s3cret"))
	a, err := ParseBasicAuthentication("Basic " + enc)
	require.NoError(t, err)
	require.Equal(t, "alice", a.User)
	require.Equal(t, "s3cret", a.Password)

	_, err = ParseBasicAuthentication("Bearer x")
	require.Error(t, err)
	_, err = ParseBasicAuthentication("Basic " + base64.StdEncoding.EncodeToString([]byte("nocolon")))
	require.Error(t, err)
	_, err = ParseBasicAuthentication("Basic " + base64.StdEncoding.EncodeToString([]byte(":pass")))
	require.Error(t, err)
}
