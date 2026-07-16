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
