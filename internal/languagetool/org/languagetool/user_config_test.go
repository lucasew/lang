package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserConfig(t *testing.T) {
	EnableABTests()
	require.True(t, HasABTestsEnabled())
	u := NewUserConfig()
	u.AddAcceptedPhrase("New York")
	require.True(t, u.AcceptsPhrase("New York"))
	u.SetConfigValueByID("RULE", []any{5})
	require.Equal(t, []any{5}, u.GetConfigValueByID("RULE"))
}
