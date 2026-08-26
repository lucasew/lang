package uk_test

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceRenamedRuleTest.java
//
// Java setup:
//   SimpleReplaceRenamedRule rule = new SimpleReplaceRenamedRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator (same wire as Soft twin). External package uk_test
// avoids import cycle with tagging/disambiguation/uk → rules/uk.

import (
	"testing"

	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// TestSimpleReplaceRenamedRule_Rule twins SimpleReplaceRenamedRuleTest.testRule —
// every active Java case with Java-visible outcomes.
func TestSimpleReplaceRenamedRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceRenamedRule(TestTools.getEnglishMessages())
	rule := rulesuk.NewSimpleReplaceRenamedRule(nil)
	// Java: new JLanguageTool(Ukrainian.DEFAULT_VARIANT)
	lt := wireUKSoftLT(t)

	// correct sentences:
	// matches = rule.match(lt.getAnalyzedSentence("Київ."));
	// assertEquals(0, matches.length);
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Київ."))))

	// matches = rule.match(lt.getAnalyzedSentence("Дніпродзержинська"));
	// assertEquals(1, matches.length);
	// assertEquals(Arrays.asList("Кам'янське", "кам'янський"), matches[0].getSuggestedReplacements());
	// assertTrue(matches[0].getMessage().contains("2016"));
	matches := rule.Match(lt.GetAnalyzedSentence("Дніпродзержинська"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Кам'янське", "кам'янський"}, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "2016")

	// matches = rule.match(lt.getAnalyzedSentence("дніпродзержинського."));
	// assertEquals(1, matches.length);
	// assertEquals(Arrays.asList("кам'янський"), matches[0].getSuggestedReplacements());
	matches = rule.Match(lt.GetAnalyzedSentence("дніпродзержинського."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"кам'янський"}, matches[0].GetSuggestedReplacements())

	// matches = rule.match(lt.getAnalyzedSentence("Червонознам'янка."));
	// assertEquals(1, matches.length);
	// assertEquals(Arrays.asList("Знам'янка", "Знаменка"), matches[0].getSuggestedReplacements());
	matches = rule.Match(lt.GetAnalyzedSentence("Червонознам'янка."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Знам'янка", "Знаменка"}, matches[0].GetSuggestedReplacements())
	// Java createRuleMatch: info contains "декомуніз" → DECOMUNIZATION_URL
	require.Contains(t, matches[0].URL, "wikipedia")

	// matches = rule.match(lt.getAnalyzedSentence("Переяслав-Хмельницький."));
	// assertEquals(1, matches.length);
	// assertEquals(Arrays.asList("Переяслав", "переяславський"), matches[0].getSuggestedReplacements());
	matches = rule.Match(lt.GetAnalyzedSentence("Переяслав-Хмельницький."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Переяслав", "переяславський"}, matches[0].GetSuggestedReplacements())
}

// TestSimpleReplaceRenamedRule_Meta twins Java id / description / Style ITS.
func TestSimpleReplaceRenamedRule_Meta(t *testing.T) {
	rule := rulesuk.NewSimpleReplaceRenamedRule(nil)
	require.Equal(t, "UK_SIMPLE_REPLACE_RENAMED", rule.GetID())
	require.Equal(t, "Пропозиція поточної назви для перейменованих власних назв", rule.GetDescription())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
}
