package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguages_Registry(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en-US"})
	require.True(t, L.IsLanguageSupported("en"))
	require.True(t, L.IsLanguageSupported("EN-US"))
	require.Equal(t, "English", L.GetLanguageForShortCode("en-US").GetName())
	require.Equal(t, "English", L.GetLanguageForShortCode("en-us").GetName())
	dyn := L.AddLanguage("Foo", "xx", "/tmp/foo.dic")
	require.Equal(t, "xx", dyn.GetShortCode())
	require.Panics(t, func() { L.AddLanguage("Bad", "yy", "/tmp/bad.txt") })
	L.ClearDynamic()
	require.False(t, L.IsLanguageSupported("xx"))
}

func TestLanguages_InvalidFormat(t *testing.T) {
	L := &Languages{}
	require.Panics(t, func() { L.IsLanguageSupported("somthing-totally-inv-alid") })
	require.Panics(t, func() { L.GetLanguageForShortCode("de-") })
	require.Panics(t, func() { L.GetLanguageForShortCode("dexx") }) // not found → panic
	require.Panics(t, func() { L.GetLanguageForShortCode("xyz-xx") })
	require.Error(t, ValidateLanguageCodeFormat("a-b-c-d"))
	require.NoError(t, ValidateLanguageCodeFormat("en-US"))
	require.NoError(t, ValidateLanguageCodeFormat("de"))
}

func TestLanguages_GetCopy(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "German", Code: "de"})
	L.Register(LanguageMeta{Name: "Demo", Code: "xx"})
	got := L.Get()
	require.Len(t, got, 1) // xx filtered
	// mutating returned slice does not affect registry
	got = append(got, LanguageMeta{Name: "X", Code: "zz"})
	require.Len(t, L.Get(), 1)
	withDemo := L.GetWithDemoLanguage()
	require.GreaterOrEqual(t, len(withDemo), 1)
}
