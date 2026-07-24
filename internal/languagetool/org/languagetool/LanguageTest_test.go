package languagetool

// Twin of LanguageTest — live tests live in language/LanguageTest_test.go
// (import cycle if language is imported from this package's tests).
import "testing"

func TestLanguage_RuleFileName(t *testing.T) {
	t.Skip("see language.TestLanguage_RuleFileName")
}
func TestLanguage_GetTranslatedName(t *testing.T) {
	t.Skip("see language.TestLanguage_GetTranslatedName")
}
func TestLanguage_GetShortNameWithVariant(t *testing.T) {
	t.Skip("see language.TestLanguage_GetShortNameWithVariant")
}
func TestLanguage_Equals(t *testing.T) {
	t.Skip("see language.TestLanguage_Equals")
}
func TestLanguage_EqualsConsiderVariantIfSpecified(t *testing.T) {
	t.Skip("see language.TestLanguage_EqualsConsiderVariantIfSpecified")
}
func TestLanguage_CreateDefaultJLanguageTool(t *testing.T) {
	// soft: default factory is NewJLanguageTool(code) in this package
	for _, code := range []string{"en-US", "de-DE", "fr"} {
		lt := NewJLanguageTool(code)
		if lt.GetLanguageCode() != code {
			t.Fatalf("got %s want %s", lt.GetLanguageCode(), code)
		}
		if len(lt.Analyze("test")) == 0 {
			t.Fatalf("empty analyze for %s", code)
		}
	}
}
