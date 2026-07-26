package fa

// Twin of languagetool-language-modules/fa/src/test/java/org/languagetool/rules/fa/WordCoherencyRuleTest.java
//
// Java setup:
//   JLanguageTool lt = new JLanguageTool(new Persian());
//   TextLevelRule rule = new WordCoherencyRule(TestTools.getMessages("fa"));
//   rule.match(Collections.singletonList(lt.getAnalyzedSentence(...)))
//
// Persian has no createDefaultTagger — getAnalyzedSentence uses PersianWordTokenizer
// + DummyTagger (surface forms). Official coherency.txt pairs only (no invent expand).
// Java WordCoherencyRule has commented-out TODO example pair only — no examples.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// analyzeFAWordCoherencySentence ports lt.getAnalyzedSentence for Persian:
// PersianWordTokenizer, untagged surface readings (no Persian tagger in Java Language).
func analyzeFAWordCoherencySentence(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, tokenizers.NewPersianWordTokenizer())
}

// TestWordCoherencyRule_Rules twins Java WordCoherencyRuleTest.testRules.
func TestWordCoherencyRule_Rules(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// Java: rule.match(Collections.singletonList(lt.getAnalyzedSentence("این یک اتاق است."))) → 0
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzeFAWordCoherencySentence("این یک اتاق است."),
	})))
	// Java: one AnalyzedSentence containing both variants (singletonList), not two sentences.
	// rule.match(Collections.singletonList(lt.getAnalyzedSentence("این یک اتاق است. این یک اطاق است."))) → 1
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzeFAWordCoherencySentence("این یک اتاق است. این یک اطاق است."),
	})))
}

// TestWordCoherencyRule_Metadata twins Java WordCoherencyRule id/description/category
// and confirms no example pairs (Java has commented-out TODO only).
func TestWordCoherencyRule_Metadata(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, "FA_WORD_COHERENCY", rule.GetID())
	require.Equal(t, "چند املا برای یک کلمه که یکی از آنها اولویت بیشتری دارد", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
}

// TestWordCoherencyRule_DataFilePairs asserts official /fa/coherency.txt bidirectional map.
func TestWordCoherencyRule_DataFilePairs(t *testing.T) {
	d := loadCoherencyData()
	// Java test pair
	require.Contains(t, d.WordMap, "اتاق")
	require.Contains(t, d.WordMap["اتاق"], "اطاق")
	require.Contains(t, d.WordMap, "اطاق")
	require.Contains(t, d.WordMap["اطاق"], "اتاق")
	// other official pairs (resource parity)
	require.Contains(t, d.WordMap["تهران"], "طهران")
	require.Contains(t, d.WordMap["امپراطور"], "امپراتور")
	require.Contains(t, d.WordMap["اطو"], "اتو")
	require.Contains(t, d.WordMap["طوس"], "توس")
	require.Contains(t, d.WordMap["طاقچه"], "تاقچه")
	require.Contains(t, d.WordMap["طاق"], "تاق")
}

// TestWordCoherencyRule_MessageAndSuggestion twins Java getMessage + AbstractWordCoherencyRule suggestion.
func TestWordCoherencyRule_MessageAndSuggestion(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// Established اتاق → later اطاق
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzeFAWordCoherencySentence("این یک اتاق است. این یک اطاق است."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "'اطاق' و 'اتاق' نباید در یک جا استفاده شوند", ms[0].GetMessage())
	require.Equal(t, 24, ms[0].GetFromPos())
	require.Equal(t, 28, ms[0].GetToPos())
	require.Equal(t, []string{"اتاق"}, ms[0].GetSuggestedReplacements())

	// reverse: established اطاق → later اتاق
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzeFAWordCoherencySentence("این یک اطاق است. این یک اتاق است."),
	})
	require.Len(t, ms, 1)
	require.Equal(t, "'اتاق' و 'اطاق' نباید در یک جا استفاده شوند", ms[0].GetMessage())
	require.Equal(t, []string{"اطاق"}, ms[0].GetSuggestedReplacements())
}

// TestWordCoherencyRule_CallIndependence: separate match calls do not share state.
func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
			analyzeFAWordCoherencySentence(s),
		})))
	}
	assertGood("این یک اتاق است.")
	assertGood("این یک اطاق است.") // not noticed across independent calls
}
