package pt

// No dedicated Java test class for PreReformPortugueseCompoundRule.
// Twin fields from PreReformPortugueseCompoundRule.java + outcomes from official list.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPreReformCompoundRule_PreReformCompounds(t *testing.T) {
	rule := NewPreReformPortugueseCompoundRule(nil)
	require.Equal(t, "PT_COMPOUNDS_PRE_REFORM", rule.GetID())

	// Official list: guarda-chuva* → only hyphenated form is correct.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("guarda-chuva"))))

	// Space form of a known * compound → one match suggesting hyphenated form.
	// getSpecificRuleId → PT_COMPOUNDS_PRE_REFORM_GUARDA_CHUVA
	matches := rule.Match(languagetool.AnalyzePlain("guarda chuva"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "PT_COMPOUNDS_PRE_REFORM_GUARDA_CHUVA", matches[0].GetSpecificRuleId())
	desc, ok := matches[0].GetRule().(interface{ GetDescription() string })
	require.True(t, ok)
	// Java getDescription is fixed "Palavras compostas" (no $match expansion).
	require.Equal(t, "Palavras compostas", desc.GetDescription())
	sugs := matches[0].GetSuggestedReplacements()
	require.Equal(t, 1, len(sugs))
	require.Equal(t, "guarda-chuva", sugs[0])

	// Sentence-start uppercase still matches (SentenceStartsWithUpperCase default true).
	assertInvalidPreReformCompound(t, rule, "Guarda chuva", "Guarda-chuva")
	assertInvalidPreReformCompound(t, rule, "abaixo assinado", "abaixo-assinado")
	assertInvalidPreReformCompound(t, rule, "bem vindo", "bem-vindo")
	assertInvalidPreReformCompound(t, rule, "Grã Bretanha", "Grã-Bretanha")
	assertInvalidPreReformCompound(t, rule, "super herói", "super-herói")

	// Hyphenated forms from the * list → 0 matches.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("abaixo-assinado"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("bem-vindo"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Grã-Bretanha"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("super-herói"))))
}

func assertInvalidPreReformCompound(t *testing.T, rule *PreReformPortugueseCompoundRule, text, suggestion string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(text))
	require.Equal(t, 1, len(matches), "text=%q", text)
	sugs := matches[0].GetSuggestedReplacements()
	require.Equal(t, 1, len(sugs), "text=%q sugs=%v", text, sugs)
	require.Equal(t, suggestion, sugs[0], "text=%q", text)
}

func TestPreReformCompoundRule_Metadata(t *testing.T) {
	rule := NewPreReformPortugueseCompoundRule(nil)

	// Java getId / getDescription
	require.Equal(t, "PT_COMPOUNDS_PRE_REFORM", rule.GetID())
	require.Equal(t, "Palavras compostas", rule.GetDescription())

	// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages))
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "COMPOUNDING", rule.GetCategory().GetID().String())

	// Java: setLocQualityIssueType(ITSIssueType.Grammar)
	require.Equal(t, rules.ITSGrammar, rule.GetLocQualityIssueType())

	// Java: no getUrl override
	require.Equal(t, "", rule.GetURL())

	// Java: useSubRuleSpecificIds()
	require.True(t, rule.SubRuleSpecificIDs)

	// Java Language short code "pt" for SpecificId
	require.Equal(t, "pt", rule.LanguageCode)

	// Java AbstractCompoundRule.sentenceStartsWithUpperCase default true
	require.True(t, rule.SentenceStartsWithUpperCase)

	// Java CompoundRuleData("/pt/pre-reform-compounds.txt")
	require.Equal(t, PreReformPortugueseCompoundFilePath, "/pt/pre-reform-compounds.txt")
	require.NotNil(t, rule.GetCompoundRuleData())

	// Java ctor messages to AbstractCompoundRule
	require.Equal(t, "Esta palavra é hifenizada.", rule.WithHyphenMessage)
	require.Equal(t, "Esta palavra é composta por justaposição.", rule.WithoutHyphenMessage)
	require.Equal(t, "Esta palavra pode ser composta por justaposição ou hifenizada.", rule.WithOrWithoutHyphenMessage)
	require.Equal(t, "Este conjunto forma uma palavra composta.", rule.ShortDesc)

	// Java: no mergeCompound override
	require.Nil(t, rule.MergeCompoundFn)
}
