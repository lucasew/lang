package ekavian

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/SimpleGrammarEkavianReplaceRuleTest.java
// Java: rule.match(lt.getAnalyzedSentence(...)) with Serbian JLanguageTool analysis.
// Go: AnalyzePlain is the untagged analysis path used by other AbstractSimpleReplaceRule twins
// when lemma tagging is not required for surface-key matches (dictionary keys are surface forms).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestSimpleGrammarEkavianReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleGrammarEkavianReplaceRule(TestTools.getEnglishMessages())
	rule := NewSimpleGrammarEkavianReplaceRule(nil)

	// Metadata twin of SimpleGrammarEkavianReplaceRule overrides + AbstractSimpleReplaceRule ctor.
	require.Equal(t, SimpleGrammarEkavianReplaceRuleID, rule.GetID())
	require.Equal(t, "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE", rule.GetID())
	require.Equal(t, "Провера граматички погрешних речи или израза", rule.GetDescription())
	require.Equal(t, "Граматички погрешна реч тј. израз", rule.GetShort())
	require.Equal(t, "sr", rule.GetLocaleCode())
	require.True(t, rule.CheckLemmas) // Java AbstractSimpleReplaceRule default
	require.True(t, rule.CaseSensitive) // Java isCaseSensitive default; SR does not override
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())
	require.Contains(t, rule.GetWrongWords(), "вишљи")
	require.Equal(t, []string{"виши"}, rule.GetWrongWords()["вишљи"])
	require.Contains(t, rule.GetWrongWords(), "сумљам")
	require.Equal(t, []string{"сумњам"}, rule.GetWrongWords()["сумљам"])

	// correct sentences:
	// Java: matches = rule.match(lt.getAnalyzedSentence("Данас је диван дан."));
	//        assertEquals(0, matches.length);
	matches := rule.Match(languagetool.AnalyzePlain("Данас је диван дан."))
	require.Equal(t, 0, len(matches))

	// incorrect sentences:
	// Java: "Син је вишљи од оца." → 1 match, suggestions ["виши"]
	matches = rule.Match(languagetool.AnalyzePlain("Син је вишљи од оца."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, []string{"виши"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, SimpleGrammarEkavianReplaceRuleID, matches[0].GetSpecificRuleId())
	require.Equal(t, "Граматички погрешна реч тј. израз", matches[0].GetShortMessage())
	// Java getMessage("вишљи", ["виши"]):
	// Не каже се „вишљи“ него „виши“.
	wantMsg := "Не каже се „вишљи“ него „виши“."
	require.Equal(t, wantMsg, matches[0].GetMessage())
	require.Equal(t, wantMsg, rule.GetMessage("вишљи", []string{"виши"}))

	// Java: "У то оправдано сумљам." → 1 match, suggestions ["сумњам"]
	matches = rule.Match(languagetool.AnalyzePlain("У то оправдано сумљам."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, []string{"сумњам"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Не каже се „сумљам“ него „сумњам“.", matches[0].GetMessage())

	// Extra surface case still in official replace-grammar.txt (еуро=евро); not in Java matrix.
	matches = rule.Match(languagetool.AnalyzePlain("Плаћа у еуро."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"евро"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Не каже се „еуро“ него „евро“.", matches[0].GetMessage())
}
