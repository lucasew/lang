package patterns

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternRuleLoaderPermission_PermissionManager(t *testing.T) {
	// PasswordAuthenticator is the permission-adjacent surface for authenticated rule sources.
	u, err := url.Parse("https://user:secret@example.com/rules.xml")
	require.NoError(t, err)
	creds, err := GetPasswordAuthenticationFromURL(u)
	require.NoError(t, err)
	require.NotNil(t, creds)
	require.Equal(t, "user", creds.Username)
	require.Equal(t, "secret", creds.Password)
	// no credentials
	u2, _ := url.Parse("https://example.com/rules.xml")
	c2, err := GetPasswordAuthenticationFromURL(u2)
	require.NoError(t, err)
	require.Nil(t, c2)
}
