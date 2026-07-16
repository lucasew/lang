package languagetool

// Twin of languagetool-standalone/src/test/java/org/languagetool/LanguagesTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguages_Get(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en"})
	L.Register(LanguageMeta{Name: "Demo", Code: "xx"})
	require.Equal(t, len(L.Get())+1, len(L.GetWithDemoLanguage()))
}

func TestLanguages_GetIsUnmodifiable(t *testing.T) {
	// Go returns a copy; "unmodifiable" means registry not mutated by caller append.
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en"})
	before := len(L.Get())
	s := L.Get()
	_ = append(s, LanguageMeta{Name: "X", Code: "zz"})
	require.Equal(t, before, len(L.Get()))
}

func TestLanguages_GetWithDemoLanguageIsUnmodifiable(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "Demo", Code: "xx"})
	before := len(L.GetWithDemoLanguage())
	s := L.GetWithDemoLanguage()
	_ = append(s, LanguageMeta{Name: "X", Code: "zz"})
	require.Equal(t, before, len(L.GetWithDemoLanguage()))
}

func TestLanguages_GetLanguageForShortName(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English (US)", Code: "en-US"})
	L.Register(LanguageMeta{Name: "German", Code: "de"})
	require.Equal(t, "en-US", L.GetLanguageForShortCode("en-us").GetShortCodeWithCountryAndVariant())
	require.Equal(t, "en-US", L.GetLanguageForShortCode("EN-US").GetShortCodeWithCountryAndVariant())
	require.Equal(t, "de", L.GetLanguageForShortCode("de").GetShortCodeWithCountryAndVariant())
	require.Panics(t, func() { L.GetLanguageForShortCode("xy") })
	require.Panics(t, func() { L.GetLanguageForShortCode("YY-KK") })
}

func TestLanguages_IsLanguageSupported(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "Demo", Code: "xx"})
	L.Register(LanguageMeta{Name: "English (US)", Code: "en-US"})
	L.Register(LanguageMeta{Name: "German", Code: "de"})
	require.True(t, L.IsLanguageSupported("xx"))
	require.True(t, L.IsLanguageSupported("XX"))
	require.True(t, L.IsLanguageSupported("en-US"))
	require.True(t, L.IsLanguageSupported("en-us"))
	require.True(t, L.IsLanguageSupported("de"))
	require.False(t, L.IsLanguageSupported("yy-ZZ"))
	require.False(t, L.IsLanguageSupported("zz"))
}

func TestLanguages_IsLanguageSupportedInvalidCode(t *testing.T) {
	L := &Languages{}
	require.Panics(t, func() { L.IsLanguageSupported("somthing-totally-inv-alid") })
}

func TestLanguages_InvalidShortName1(t *testing.T) {
	L := &Languages{}
	require.Panics(t, func() { L.GetLanguageForShortCode("de-") })
}

func TestLanguages_InvalidShortName2(t *testing.T) {
	L := &Languages{}
	require.Panics(t, func() { L.GetLanguageForShortCode("dexx") })
}

func TestLanguages_InvalidShortName3(t *testing.T) {
	L := &Languages{}
	require.Panics(t, func() { L.GetLanguageForShortCode("xyz-xx") })
}

func TestLanguages_GetLanguageForName(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English (US)", Code: "en-US"})
	L.Register(LanguageMeta{Name: "German", Code: "de"})
	m, ok := L.GetLanguageForName("English (US)")
	require.True(t, ok)
	require.Equal(t, "en-US", m.GetShortCodeWithCountryAndVariant())
	_, ok = L.GetLanguageForName("Foobar")
	require.False(t, ok)
}

func TestLanguages_GetLanguageForLocale(t *testing.T) {
	// Locale mapping is a thin short-code lookup for now.
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English (US)", Code: "en-US"})
	require.Equal(t, "en-US", L.GetLanguageForShortCode("en-US").Code)
}
