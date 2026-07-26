package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PostReformCompoundRuleTest.java
// Surface AnalyzePlain path (not full JLanguageTool picky mode).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPostReformCompoundRule_PostReformCompounds(t *testing.T) {
	rule := NewPostReformPortugueseCompoundRule(nil)
	require.Equal(t, "PT_COMPOUNDS_POST_REFORM", rule.GetID())

	// Java: assertValidCompound("super-herói");
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("super-herói"))))

	// Java: checkCompound("super herói").size() == 1
	// getSpecificRuleId → PT_COMPOUNDS_POST_REFORM_SUPER_HERÓI
	// getRule().getDescription() → Erro na formação da palavra composta "super herói"
	matches := rule.Match(languagetool.AnalyzePlain("super herói"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "PT_COMPOUNDS_POST_REFORM_SUPER_HERÓI", matches[0].GetSpecificRuleId())
	desc, ok := matches[0].GetRule().(interface{ GetDescription() string })
	require.True(t, ok)
	require.Equal(t, `Erro na formação da palavra composta "super herói"`, desc.GetDescription())

	// Java: assertInvalidCompound(text, suggestion)
	assertInvalidCompound(t, rule, "Super estrela", "Superestrela")
	assertInvalidCompound(t, rule, "web-site", "website")
	assertInvalidCompound(t, rule, "Grã Bretanha", "Grã-Bretanha")
	assertInvalidCompound(t, rule, "ultra-som", "ultrassom")
	assertInvalidCompound(t, rule, "ultra-realismo", "ultrarrealismo")
	assertInvalidCompound(t, rule, "anti semita", "antissemita")
	assertInvalidCompound(t, rule, "arqui-rabino", "arquirrabino")
	assertInvalidCompound(t, rule, "ópera rock", "ópera-rock")
}

func assertInvalidCompound(t *testing.T, rule *PostReformPortugueseCompoundRule, text, suggestion string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(text))
	require.Equal(t, 1, len(matches), "text=%q", text)
	sugs := matches[0].GetSuggestedReplacements()
	require.Equal(t, 1, len(sugs), "text=%q sugs=%v", text, sugs)
	require.Equal(t, suggestion, sugs[0], "text=%q", text)
}

func TestPostReformCompoundRule_Metadata(t *testing.T) {
	rule := NewPostReformPortugueseCompoundRule(nil)

	// Java getId / getDescription
	require.Equal(t, "PT_COMPOUNDS_POST_REFORM", rule.GetID())
	require.Equal(t, `Erro na formação da palavra composta "$match"`, rule.GetDescription())

	// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages))
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "COMPOUNDING", rule.GetCategory().GetID().String())

	// Java: setLocQualityIssueType(ITSIssueType.Grammar)
	require.Equal(t, rules.ITSGrammar, rule.GetLocQualityIssueType())

	// Java: getUrl → Wikipedia AO 1990 list
	require.Equal(t, PostReformPortugueseCompoundURL, rule.GetURL())

	// Java: useSubRuleSpecificIds()
	require.True(t, rule.SubRuleSpecificIDs)

	// Java Language short code "pt" for SpecificId
	require.Equal(t, "pt", rule.LanguageCode)

	// Java CompoundRuleData("/pt/post-reform-compounds.txt")
	require.Equal(t, PostReformPortugueseCompoundFilePath, "/pt/post-reform-compounds.txt")
	require.NotNil(t, rule.GetCompoundRuleData())

	// Java ctor messages to AbstractCompoundRule
	require.Equal(t, "Esta palavra é hifenizada.", rule.WithHyphenMessage)
	require.Equal(t, "Esta palavra é composta por justaposição.", rule.WithoutHyphenMessage)
	require.Equal(t, "Esta palavra pode ser composta por justaposição ou hifenizada.", rule.WithOrWithoutHyphenMessage)
	require.Equal(t, "Este conjunto forma uma palavra composta.", rule.ShortDesc)

	// Java mergeCompound override wired
	require.NotNil(t, rule.MergeCompoundFn)
}

func TestPostReformMergeCompound_Digraph(t *testing.T) {
	// Direct twin of Java mergeCompound digraph behaviour
	require.Equal(t, "ultrassom", postReformMergeCompound("ultra-som", true))
	require.Equal(t, "ultrarrealismo", postReformMergeCompound("ultra-realismo", true))
	require.Equal(t, "antissemita", postReformMergeCompound("anti semita", true))
	require.Equal(t, "arquirrabino", postReformMergeCompound("arqui-rabino", true))
	// no digraph when previous does not end vowel / current not r|s
	require.Equal(t, "website", postReformMergeCompound("web-site", true))
	require.Equal(t, "Superestrela", postReformMergeCompound("Super estrela", true))
}
