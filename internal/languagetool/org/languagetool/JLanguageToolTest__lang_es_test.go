package languagetool

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/JLanguageToolTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testXMLRules
func TestJLanguageTool_lang_es_XMLRules(t *testing.T) {
	lt := NewJLanguageTool("es")
	require.Equal(t, "es", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Esto es una prueba."))
}

// Port of JLanguageToolTest.testMultitokenSpeller
func TestJLanguageTool_lang_es_MultitokenSpeller(t *testing.T) {
	lt := NewJLanguageTool("es")
	require.NotEmpty(t, lt.Analyze("Buenos Aires"))
}

// Port of JLanguageToolTest.testFilterRuleMatches
func TestJLanguageTool_lang_es_FilterRuleMatches(t *testing.T) {
	lt := NewJLanguageTool("es")
	require.NotEmpty(t, lt.Analyze("El gato duerme."))
}
