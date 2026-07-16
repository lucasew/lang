package languagetool

// Twin of languagetool-standalone/src/test/java/org/languagetool/LanguageTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-standalone/src/test/java/org/languagetool/LanguageTest.java :: LanguageTest.testRuleFileName
func TestLanguage_RuleFileName(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertTrue
}

// Port of languagetool-standalone/src/test/java/org/languagetool/LanguageTest.java :: LanguageTest.testGetTranslatedName
func TestLanguage_GetTranslatedName(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-standalone/src/test/java/org/languagetool/LanguageTest.java :: LanguageTest.testGetShortNameWithVariant
func TestLanguage_GetShortNameWithVariant(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-standalone/src/test/java/org/languagetool/LanguageTest.java :: LanguageTest.testEquals
func TestLanguage_Equals(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-standalone/src/test/java/org/languagetool/LanguageTest.java :: LanguageTest.testEqualsConsiderVariantIfSpecified
func TestLanguage_EqualsConsiderVariantIfSpecified(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
}

// Port of languagetool-standalone/src/test/java/org/languagetool/LanguageTest.java :: LanguageTest.testCreateDefaultJLanguageTool
func TestLanguage_CreateDefaultJLanguageTool(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}
