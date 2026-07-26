package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianWordRepeatRuleTest.java
//
// Java setup:
//   final RussianWordRepeatRule rule = new RussianWordRepeatRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(Russian.getInstance());
//   matches = rule.match(lt.getAnalyzedSentence(...));
//
// Go: NewRussianWordRepeatRule + NewJLanguageTool("ru") with RussianTagger (real russian.dict).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/stretchr/testify/require"
)

// wireRUWordRepeatLT ports JLanguageTool(Russian.getInstance()) for this rule twin.
// Skips when russian.dict is unavailable.
func wireRUWordRepeatLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	if tagru.DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	tagru.EnsureDefaultRussianTagger()
	require.NotNil(t, tagru.DefaultRussianTagger)
	require.NotNil(t, tagru.DefaultRussianTagger.GetWordTagger())
	require.NotEmpty(t, tagru.RussianPOSDictPath(), "real russian.dict must load")

	lt := languagetool.NewJLanguageTool("ru")
	// Java: language.getTagger().tag(tokens) — full-sentence RussianTagger.Tag path.
	lt.SentenceTagger = tagru.DefaultRussianTagger.Tag
	return lt
}

// TestRussianWordRepeatRule_Rule twins Java RussianWordRepeatRuleTest.testRule.
func TestRussianWordRepeatRule_Rule(t *testing.T) {
	// Java: new RussianWordRepeatRule(TestTools.getEnglishMessages())
	rule := NewRussianWordRepeatRule(nil)
	lt := wireRUWordRepeatLT(t)

	// correct
	matches := rule.Match(lt.GetAnalyzedSentence("Повтор слов в предложении."))
	require.Equal(t, 0, len(matches))

	// incorrect
	matches = rule.Match(lt.GetAnalyzedSentence("Повтор слов в повтор предложении."))
	require.Equal(t, 1, len(matches))
}

// Metadata twin (Java getId / getDescription / getMessage / getShortMessage /
// AdvancedWordRepeatRule ctor: MISC, defaultOff, Style + example pair).
func TestRussianWordRepeatRule_Metadata(t *testing.T) {
	r := NewRussianWordRepeatRule(nil)
	require.Equal(t, "RU_WORD_REPEAT", r.GetID())
	require.Equal(t, "Повтор слов в предложении", r.GetDescription())
	require.Equal(t, "Повтор слов в предложении", r.Message)
	require.Equal(t, "Повтор слов в предложении", r.ShortMessage)
	require.True(t, r.IsDefaultOff())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(rules.CategoryMisc))
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.Equal(t, []string{"доме"}, r.GetIncorrectExamples()[0].GetCorrections())
}
