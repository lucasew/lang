package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseColourHyphenationRuleTest.java
// Java uses full JLanguageTool pt-BR with only PT_COLOUR_HYPHENATION enabled (PICKY).
// Go Match(AnalyzePlain(...)) when outcomes match Java-visible results (counts + specific id).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPortugueseColourHyphenationRule_Test(t *testing.T) {
	rule := NewPortugueseColourHyphenationRule(nil)
	require.Equal(t, "PT_COLOUR_HYPHENATION", rule.GetID())

	// Java: assertThat(checkCompound("azul-claro").size(), is(0));
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("azul-claro"))))

	// Java: assertThat(checkCompound("azul claro").size(), is(1));
	matches := rule.Match(languagetool.AnalyzePlain("azul claro"))
	require.Equal(t, 1, len(matches))

	// Java: getSpecificRuleId() → "PT_COLOUR_HYPHENATION_AZUL_CLARO"
	require.Equal(t, "PT_COLOUR_HYPHENATION_AZUL_CLARO", matches[0].GetSpecificRuleId())
}

func TestPortugueseColourHyphenationRule_Metadata(t *testing.T) {
	rule := NewPortugueseColourHyphenationRule(nil)

	// Java getId / getDescription
	require.Equal(t, "PT_COLOUR_HYPHENATION", rule.GetID())
	require.Equal(t, "Nomes de cores devem ser hifenizados: \"$match\"", rule.GetDescription())

	// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages))
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "COMPOUNDING", rule.GetCategory().GetID().String())

	// Java: setLocQualityIssueType(ITSIssueType.Grammar)
	require.Equal(t, rules.ITSGrammar, rule.GetLocQualityIssueType())

	// Java: getUrl → Wikipedia AO 1990 list
	require.Equal(t, PortugueseColourHyphenationURL, rule.GetURL())

	// Java: useSubRuleSpecificIds()
	require.True(t, rule.SubRuleSpecificIDs)

	// Java CompoundRuleData("/pt/compound_colours.txt")
	require.Equal(t, PortugueseColourHyphenationFilePath, "/pt/compound_colours.txt")
	require.NotNil(t, rule.GetCompoundRuleData())

	// Java ctor messages to AbstractCompoundRule
	require.Equal(t, "Nomes de cores são palavras compostas e devem ser hifenizados.", rule.WithHyphenMessage)
	require.Equal(t, "Esta palavra é composta por justaposição.", rule.WithoutHyphenMessage)
	require.Equal(t, "Esta palavra pode ser composta por justaposição ou hifenizada.", rule.WithOrWithoutHyphenMessage)
	require.Equal(t, "Nomes de cores são palavras compostas.", rule.ShortDesc)
}
