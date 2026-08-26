package languagetool

// Twin of languagetool-language-modules/eo/src/test/java/org/languagetool/EsperantoTest.java
//
// Java:
//   Esperanto language = new Esperanto();
//   assertThat(language.getCountries().length, is(0));
//   JLanguageTool languageTool = new JLanguageTool(language);
//   String input = "La Mondaj Ludoj de 2013 estis plur-sporta evento...";
//   List<RuleMatch> ruleMatches = languageTool.check(input);
//   assertThat(ruleMatches.size(), is(1));
//   assertThat(ruleMatches.get(0).getRule().getId(), is("HUNSPELL_RULE"));
//
// Full metadata (empty countries, HUNSPELL_RULE on sample) lives in language/ and rules/eo/
// twins — this package cannot import language/ or rules/eo without an import cycle.
// Assert Analyze surface on the Java sample text here.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.EsperantoTest.test — Analyze surface (metadata + HUNSPELL
// outcomes in language/esperanto_test.go and rules/eo/EsperantoTest_test.go).
func TestEsperanto_Test(t *testing.T) {
	lt := NewJLanguageTool("eo")
	require.Equal(t, "eo", lt.GetLanguageCode())
	// Java sample used for check → HUNSPELL_RULE
	input := "La Mondaj Ludoj de 2013 estis plur-sporta evento..."
	require.NotEmpty(t, lt.Analyze(input))
}
