package language

// Twin of languagetool-core/src/test/java/org/languagetool/language/LanguageBuilderTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageBuilder_MakeAdditionalLanguage(t *testing.T) {
	meta, err := MakeAdditionalLanguage("rules-xy-Fakelanguage.xml")
	require.NoError(t, err)
	require.Equal(t, "Fakelanguage", meta.Name)
	require.Equal(t, "xy", meta.Code)
	require.True(t, meta.Additional)
	require.Equal(t, "rules-xy-Fakelanguage.xml", meta.RulesFile)
}

func TestLanguageBuilder_IllegalFileName(t *testing.T) {
	_, err := MakeAdditionalLanguage("foo")
	require.Error(t, err)
	var rfe *RuleFilenameException
	require.ErrorAs(t, err, &rfe)
}

func TestLanguageBuilder_ExtendedLanguage(t *testing.T) {
	// de is typically registered; ExtendedLanguage adds rule file
	ext, err := MakeExtendedLanguage("rules-de-Custom.xml", []string{"/base/grammar.xml"})
	require.NoError(t, err)
	require.Equal(t, "Custom", ext.GetName())
	require.True(t, ext.IsExternal())
	require.Equal(t, []string{"/base/grammar.xml", "rules-de-Custom.xml"}, ext.GetRuleFileNames())
	require.Equal(t, "de", ext.GetShortCode())
}

func TestLanguageBuilder_CodeParts(t *testing.T) {
	require.Equal(t, "en", ShortCodeFromParts("en_US"))
	require.Equal(t, []string{"US"}, CountriesFromParts("en_US"))
	require.Equal(t, []string{""}, CountriesFromParts("de"))
}

func TestLanguageBuilder_ThreePartsOnly(t *testing.T) {
	_, err := MakeAdditionalLanguage("rules-en-My-Lang.xml") // 4 parts
	require.Error(t, err)
}
