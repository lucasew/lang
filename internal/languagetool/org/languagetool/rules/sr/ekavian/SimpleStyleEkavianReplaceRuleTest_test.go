package ekavian

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/SimpleStyleEkavianReplaceRuleTest.java
// Java: rule.match(lt.getAnalyzedSentence(...)) with Serbian JLanguageTool analysis.
// Go: AnalyzePlain is the untagged analysis path used by other AbstractSimpleReplaceRule twins
// when lemma tagging is not required for surface-key matches (dictionary keys are surface forms).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestSimpleStyleEkavianReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleStyleEkavianReplaceRule(TestTools.getEnglishMessages())
	rule := NewSimpleStyleEkavianReplaceRule(nil)

	// Metadata twin of SimpleStyleEkavianReplaceRule overrides + AbstractSimpleReplaceRule ctor.
	require.Equal(t, SimpleStyleEkavianReplaceRuleID, rule.GetID())
	require.Equal(t, "SR_EKAVIAN_SIMPLE_STYLE_REPLACE_RULE", rule.GetID())
	require.Equal(t, "Провера стилски лоших речи или израза", rule.GetDescription())
	require.Equal(t, "Стилски лоша реч тј. израз", rule.GetShort())
	require.Equal(t, "sr", rule.GetLocaleCode())
	require.True(t, rule.CheckLemmas)   // Java AbstractSimpleReplaceRule default
	require.True(t, rule.CaseSensitive) // Java isCaseSensitive default; SR does not override
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())
	require.Contains(t, rule.GetWrongWords(), "дебела")
	require.Equal(t, []string{"елегантно попуњена"}, rule.GetWrongWords()["дебела"])
	require.Contains(t, rule.GetWrongWords(), "компјутер")
	require.Equal(t, []string{"рачунар"}, rule.GetWrongWords()["компјутер"])

	// correct sentences:
	// Java: matches = rule.match(lt.getAnalyzedSentence("Он је добар."));
	//        assertEquals(0, matches.length);
	matches := rule.Match(languagetool.AnalyzePlain("Он је добар."))
	require.Equal(t, 0, len(matches))

	// incorrect sentences:
	// Java: "Она је дебела." → 1 match, suggestions ["елегантно попуњена"]
	matches = rule.Match(languagetool.AnalyzePlain("Она је дебела."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, []string{"елегантно попуњена"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, SimpleStyleEkavianReplaceRuleID, matches[0].GetSpecificRuleId())
	require.Equal(t, "Стилски лоша реч тј. израз", matches[0].GetShortMessage())
	// Java getMessage("дебела", ["елегантно попуњена"]):
	// Уместо израза „дебела“ било би боље да користите: елегантно попуњена.
	wantMsg := "Уместо израза „дебела“ било би боље да користите: елегантно попуњена."
	require.Equal(t, wantMsg, matches[0].GetMessage())
	require.Equal(t, wantMsg, rule.GetMessage("дебела", []string{"елегантно попуњена"}))

	// Extra surface case still in official replace-style.txt (компјутер=рачунар); not in Java matrix.
	matches = rule.Match(languagetool.AnalyzePlain("Купио је компјутер."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"рачунар"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Уместо израза „компјутер“ било би боље да користите: рачунар.", matches[0].GetMessage())
}
