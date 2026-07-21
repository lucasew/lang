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

// Ports Java Languages class-init: built-in modules available for canLanguageBeDetected.
func TestEnsureBuiltInLanguagesRegistered(t *testing.T) {
	EnsureBuiltInLanguagesRegistered()
	require.True(t, GlobalLanguages.IsLanguageSupported("en"))
	require.True(t, GlobalLanguages.IsLanguageSupported("en-US"))
	require.True(t, GlobalLanguages.IsLanguageSupported("de-DE"))
	require.True(t, GlobalLanguages.IsLanguageSupported("fr"))
	// multi-country Spanish → longCode is bare "es"
	es := GlobalLanguages.GetLanguageForShortCode("es")
	require.Equal(t, "es", es.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "Spanish", es.GetName())
	// single-country AmericanEnglish
	enUS := GlobalLanguages.GetLanguageForShortCode("en-US")
	require.Equal(t, "English (US)", enUS.GetName())
	require.Equal(t, "en", enUS.GetShortCode())
	require.Equal(t, "en-US", enUS.GetShortCodeWithCountryAndVariant())
	require.True(t, enUS.IsVariant())
	require.False(t, GlobalLanguages.GetLanguageForShortCode("en").IsVariant())
	// SpanishVoseo: only AR active in Java (commented countries not in array)
	voseo := GlobalLanguages.GetLanguageForShortCode("es-AR")
	require.Equal(t, "Spanish (voseo)", voseo.GetName())
	require.Equal(t, "es-AR", voseo.GetShortCodeWithCountryAndVariant())
	require.True(t, voseo.IsVariant())
	// Catalan base longCode is ca-ES (countries=["ES"]); not a variant (extends Language)
	ca := GlobalLanguages.GetLanguageForShortCode("ca-ES")
	require.Equal(t, "Catalan", ca.GetName())
	require.Equal(t, "ca-ES", ca.GetShortCodeWithCountryAndVariant())
	require.False(t, ca.IsVariant())
	require.True(t, GlobalLanguages.HasVariant(ca))
	require.False(t, GlobalLanguages.IsHiddenFromGui(ca)) // default variant of itself
	val := GlobalLanguages.GetLanguageForShortCode("ca-ES-valencia")
	require.Equal(t, "Catalan (Valencian)", val.GetName())
	require.True(t, val.IsVariant())
	// long-code mapping for LibreOffice (fr-FR → French)
	m := GlobalLanguages.GetLongCodeToLangMapping()
	require.Equal(t, "fr", m["fr-FR"].GetShortCode())
	require.Equal(t, "French", m["fr-FR"].GetName())
	// getLanguageForShortCode uses mapping; isLanguageSupported does not
	require.Equal(t, "French", GlobalLanguages.GetLanguageForShortCode("fr-FR").GetName())
	require.False(t, GlobalLanguages.IsLanguageSupported("fr-FR"),
		"Java isLanguageSupported uses OrNull only — no long-code mapping")
	// Arabic first country is "" → no ar-SA mapping from Arabic itself
	_, arMapped := m["ar-SA"]
	require.False(t, arMapped, "Arabic countries[0] is empty — no ar-SA map entry")
	// Portuguese first country "" → no pt-CV from base Portuguese
	_, ptMapped := m["pt-CV"]
	require.False(t, ptMapped)
	// SimpleGerman force isVariant
	sg := GlobalLanguages.GetLanguageForShortCode("de-DE-x-simple-language")
	require.Equal(t, "Simple German", sg.GetName())
	require.True(t, sg.IsVariant())
	// zz/xx not registered as normal modules
	require.False(t, GlobalLanguages.IsLanguageSupported("xx"))
	// idempotent
	EnsureBuiltInLanguagesRegistered()
	require.True(t, GlobalLanguages.IsLanguageSupported("en"))
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

func TestLanguages_NoopCodes(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en-US"})
	m := L.GetLanguageForShortCodeWithNoop("tl", []string{"tl"})
	require.Equal(t, NoopLanguageCode, m.GetShortCode())
	require.Panics(t, func() { L.GetLanguageForShortCodeWithNoop("xx-YY", nil) })
	require.True(t, HasPremiumClass("org.languagetool.language.English"))
	require.False(t, HasPremiumClass("org.languagetool.language.Polish"))
	codes := L.GetLangCodes()
	require.Contains(t, codes, "en-US")
}

func TestLanguages_GetOrAddByClassName(t *testing.T) {
	L := &Languages{}
	L.Register(LanguageMeta{Name: "English", Code: "en"})
	m := L.GetOrAddLanguageByClassName("org.languagetool.language.English")
	require.Equal(t, "en", m.Code)
}
