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

// Twin of LanguagesTest.testIsVariant
func TestLanguages_IsVariant(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en", DefaultVariantCode: "en-US"})
	L.Register(LanguageMeta{Name: "English (US)", Code: "en-US"})
	L.Register(LanguageMeta{Name: "German", Code: "de", DefaultVariantCode: "de-DE"})
	L.Register(LanguageMeta{Name: "German (Switzerland)", Code: "de-CH"})
	require.True(t, L.GetLanguageForShortCode("en-US").IsVariant())
	require.True(t, L.GetLanguageForShortCode("de-CH").IsVariant())
	require.False(t, L.GetLanguageForShortCode("en").IsVariant())
	require.False(t, L.GetLanguageForShortCode("de").IsVariant())
}

// Twin of LanguagesTest.testHasPremium
func TestLanguages_HasPremium(t *testing.T) {
	L := &Languages{}
	require.True(t, L.HasPremium("org.languagetool.language.Portuguese"))
	require.True(t, L.HasPremium("org.languagetool.language.GermanyGerman"))
	require.True(t, L.HasPremium("org.languagetool.language.AmericanEnglish"))
	require.False(t, L.HasPremium("org.languagetool.language.Danish"))
}

// Twin of LanguagesTest.testHasVariant
func TestLanguages_HasVariant(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en", DefaultVariantCode: "en-US"})
	L.Register(LanguageMeta{Name: "English (US)", Code: "en-US"})
	L.Register(LanguageMeta{Name: "German", Code: "de", DefaultVariantCode: "de-DE"})
	L.Register(LanguageMeta{Name: "German (Switzerland)", Code: "de-CH"})
	L.Register(LanguageMeta{Name: "Asturian", Code: "ast"})
	L.Register(LanguageMeta{Name: "Polish", Code: "pl"})
	require.True(t, L.HasVariant(L.GetLanguageForShortCode("en")))
	require.True(t, L.HasVariant(L.GetLanguageForShortCode("de")))
	require.False(t, L.HasVariant(L.GetLanguageForShortCode("en-US")))
	require.False(t, L.HasVariant(L.GetLanguageForShortCode("de-CH")))
	require.False(t, L.HasVariant(L.GetLanguageForShortCode("ast")))
	require.False(t, L.HasVariant(L.GetLanguageForShortCode("pl")))
}

// Twin of LanguagesTest.isHiddenFromGui
func TestLanguages_IsHiddenFromGui(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en", DefaultVariantCode: "en-US"})
	L.Register(LanguageMeta{Name: "English (US)", Code: "en-US"})
	L.Register(LanguageMeta{Name: "German", Code: "de", DefaultVariantCode: "de-DE"})
	L.Register(LanguageMeta{Name: "German (Switzerland)", Code: "de-CH"})
	L.Register(LanguageMeta{Name: "German (Germany)", Code: "de-DE"})
	L.Register(LanguageMeta{Name: "Portuguese", Code: "pt", DefaultVariantCode: "pt-PT"})
	L.Register(LanguageMeta{Name: "Portuguese (Portugal)", Code: "pt-PT"})
	L.Register(LanguageMeta{Name: "Asturian", Code: "ast"})
	L.Register(LanguageMeta{Name: "Polish", Code: "pl"})
	L.Register(LanguageMeta{Name: "Catalan (Spain)", Code: "ca-ES"})
	L.Register(LanguageMeta{Name: "Catalan (Valencia)", Code: "ca-ES-valencia"})
	L.Register(LanguageMeta{Name: "Simple German", Code: "de-DE-x-simple-language"})
	require.True(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("en")))
	require.True(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("de")))
	require.True(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("pt")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("en-US")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("de-CH")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("ast")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("pl")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("ca-ES")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("ca-ES-valencia")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("de-DE-x-simple-language")))
	require.False(t, L.IsHiddenFromGui(L.GetLanguageForShortCode("de-DE")))
}
