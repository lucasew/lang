package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserConfig(t *testing.T) {
	EnableABTests()
	require.True(t, HasABTestsEnabled())
	u := NewUserConfig()
	require.Equal(t, "default", u.GetUserDictName())
	require.True(t, u.IsSuggestionsEnabled())
	require.Equal(t, TokenNone, u.TokenType)
	u.AddAcceptedPhrase("New York")
	require.True(t, u.AcceptsPhrase("New York"))
	u.SetConfigValueByID("RULE", []any{5})
	require.Equal(t, []any{5}, u.GetConfigValueByID("RULE"))
	require.Nil(t, u.GetConfigValueByID("MISSING"))
}

func TestUserConfig_AcceptedPhrasesFromWords(t *testing.T) {
	u := NewUserConfigWithWords([]string{"foo", "New York", "bar baz"}, nil)
	require.True(t, u.AcceptsPhrase("New York"))
	require.True(t, u.AcceptsPhrase("bar baz"))
	require.False(t, u.AcceptsPhrase("foo"))
	require.Equal(t, []string{"foo", "New York", "bar baz"}, u.GetAcceptedWords())
}

func TestUserConfig_PreferredLanguages(t *testing.T) {
	u := NewUserConfig()
	u.SetPreferredLanguagesList([]string{"ja", "en", "de", "xx", "fr"})
	// main langs sorted, >=2 → join
	require.Equal(t, "de,en,fr", u.PreferredLanguages)
	require.Equal(t, []string{"de", "en", "fr"}, u.GetPreferredLanguages())
	u.SetPreferredLanguagesList([]string{"en"})
	require.Equal(t, "", u.PreferredLanguages)
	// Java "".split(",") → [""]
	require.Equal(t, []string{""}, u.GetPreferredLanguages())
}

func TestUserConfig_Equal(t *testing.T) {
	a := NewUserConfigWithWords([]string{"a"}, map[string][]any{"R": {1}})
	b := NewUserConfigWithWords([]string{"a"}, map[string][]any{"R": {1}})
	require.True(t, a.Equal(b))
	b.UserSpecificSpellerWords = []string{"b"}
	require.False(t, a.Equal(b))
	require.Contains(t, a.String(), "dictionarySize=1")
}
