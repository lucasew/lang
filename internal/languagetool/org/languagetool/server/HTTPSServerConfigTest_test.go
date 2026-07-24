package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/HTTPSServerConfigTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of HTTPSServerConfigTest.testArgumentParsing
func TestHTTPSServerConfig_ArgumentParsing(t *testing.T) {
	c := NewHTTPSServerConfigPort(8443, true, "/tmp/ks.jks", "secret")
	require.Equal(t, 8443, c.Port)
	require.True(t, c.Verbose)
	require.Equal(t, "/tmp/ks.jks", c.GetKeystore())
	require.Equal(t, "secret", c.GetKeyStorePassword())
}

// Port of HTTPSServerConfigTest.testMinimalPropertyFile
func TestHTTPSServerConfig_MinimalPropertyFile(t *testing.T) {
	c := NewHTTPSServerConfig("", "")
	err := c.ApplyKeystoreProps(map[string]string{
		"keystore": "/path/to/keystore.jks",
		"password": "changeit",
	})
	require.NoError(t, err)
	require.Equal(t, "/path/to/keystore.jks", c.GetKeystore())
	require.Equal(t, "changeit", c.GetKeyStorePassword())
}

// Port of HTTPSServerConfigTest.testMissingPropertyFile
func TestHTTPSServerConfig_MissingPropertyFile(t *testing.T) {
	c := NewHTTPSServerConfig("", "")
	err := c.ApplyKeystoreProps(nil)
	require.Error(t, err)
	var ice *IllegalConfigurationError
	require.ErrorAs(t, err, &ice)
}

// Port of HTTPSServerConfigTest.testIncompletePropertyFile
func TestHTTPSServerConfig_IncompletePropertyFile(t *testing.T) {
	c := NewHTTPSServerConfig("", "")
	err := c.ApplyKeystoreProps(map[string]string{"keystore": "/ks.jks"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "password")
	err = c.ApplyKeystoreProps(map[string]string{"password": "x"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "keystore")
}
