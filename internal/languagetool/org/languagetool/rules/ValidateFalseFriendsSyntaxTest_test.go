package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

func TestValidateFalseFriendsSyntax_FalseFriendsXML(t *testing.T) {
	require.NoError(t, tools.ValidateFalseFriendsXML(strings.NewReader(
		`<rules lang="en"><rulegroup id="A"/></rules>`)))
	require.Error(t, tools.ValidateFalseFriendsXML(strings.NewReader(`<rules><unclosed>`)))
}
