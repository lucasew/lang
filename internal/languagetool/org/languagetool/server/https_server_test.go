package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPSServerConfig(t *testing.T) {
	c := NewHTTPSServerConfigPort(8443, true, "/path/to.jks", "secret")
	require.Equal(t, 8443, c.Port)
	require.True(t, c.Verbose)
	require.Equal(t, "/path/to.jks", c.GetKeystore())
	require.Equal(t, "secret", c.GetKeyStorePassword())

	require.Error(t, c.ApplyKeystoreProps(map[string]string{}))
	require.NoError(t, c.ApplyKeystoreProps(map[string]string{"keystore": "a.jks", "password": "p"}))
	require.Equal(t, "a.jks", c.KeystorePath)

	s := NewHTTPSServer(c, false, "localhost", DefaultAllowedIPs)
	require.Equal(t, "https", s.Protocol())
	require.True(t, s.HasKeystore())
}
