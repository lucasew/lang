package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguages_Registry(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en-US"})
	require.True(t, L.IsLanguageSupported("en"))
	require.Equal(t, "English", L.GetLanguageForShortCode("en-US").GetName())
	dyn := L.AddLanguage("Foo", "xx", "/tmp/foo.dic")
	require.Equal(t, "xx", dyn.GetShortCode())
	require.Panics(t, func() { L.AddLanguage("Bad", "yy", "/tmp/bad.txt") })
	L.ClearDynamic()
	require.False(t, L.IsLanguageSupported("xx"))
}
