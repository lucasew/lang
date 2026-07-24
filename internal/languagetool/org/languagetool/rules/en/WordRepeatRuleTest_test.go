package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/WordRepeatRuleTest.java
//
// Java setup:
//   Language english = Languages.getLanguageForShortCode("en");
//   WordRepeatRule rule = new WordRepeatRule(TestTools.getEnglishMessages(), english);
//   JLanguageTool lt = new JLanguageTool(english);
//   assertMatches: rule.match(lt.getAnalyzedSentence(input)).length == expected
//
// Uses core rules.WordRepeatRule (org.languagetool.rules.WordRepeatRule), not EnglishWordRepeatRule.
// Go analysis path: tagging/en.AnalyzeEnglishSentence — EnglishWordTokenizer +
// EnglishTagger (english.dict) + hybrid disambiguator (same resources Java loads).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/stretchr/testify/require"
)

func wordRepeatRuleENMessages() map[string]string {
	// Mirror TestTools English ResourceBundle keys used by WordRepeatRule.
	return map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition":       "Word repetition",
		"desc_repetition_short": "repetition",
		"category_misc":         "Miscellaneous",
	}
}

func wordRepeatRuleENLT(t *testing.T) {
	t.Helper()
	// Ensure english.dict is discoverable (third_party/english-pos-dict or LANG_ENGLISH_DICT).
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree (third_party/english-pos-dict or inspiration)")
	}
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, tagen.EnglishTagWord(), "EnglishTagWord must load from english.dict")
}

// assertWRRMatches ports WordRepeatRuleTest.assertMatches:
// rule.match(lt.getAnalyzedSentence(input)).length == expectedMatches
func assertWRRMatches(t *testing.T, rule *rules.WordRepeatRule, input string, expectedMatches int) {
	t.Helper()
	sent := tagen.AnalyzeEnglishSentence(input)
	require.NotNil(t, sent, "AnalyzeEnglishSentence(%q)", input)
	matches := rule.Match(sent)
	require.Equal(t, expectedMatches, len(matches), "assertMatches %q", input)
}

// TestWordRepeatRule_Rule ports WordRepeatRuleTest.testRule (EN module).
// Cases and expected counts are exact Java twins.
func TestWordRepeatRule_Rule(t *testing.T) {
	wordRepeatRuleENLT(t)
	// Java: new WordRepeatRule(TestTools.getEnglishMessages(), english)
	// Language ctor arg is unused by core WordRepeatRule.match.
	rule := rules.NewWordRepeatRule(wordRepeatRuleENMessages())
	require.Equal(t, "WORD_REPEAT_RULE", rule.GetID())

	assertWRRMatches(t, rule, "This is a test sentence.", 0)
	assertWRRMatches(t, rule, "This is a test sentence...", 0)

	// make sure we ignore immunized tokens
	assertWRRMatches(t, rule, "And side to side and top to bottom...", 0)

	assertWRRMatches(t, rule, "This this is a test sentence.", 1)
	assertWRRMatches(t, rule, "This is a test sentence sentence.", 1)
	assertWRRMatches(t, rule, "This is is a a test sentence sentence.", 3)
}

// Java createRuleMatch outcomes for a clear false-alarm: span + suggestion.
func TestWordRepeatRule_MatchSpanAndSuggestion(t *testing.T) {
	wordRepeatRuleENLT(t)
	rule := rules.NewWordRepeatRule(wordRepeatRuleENMessages())
	sent := tagen.AnalyzeEnglishSentence("This this is a test sentence.")
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	m := matches[0]
	// "This this" — from start of first This to end of second this (UTF-16 len of prevToken).
	require.Equal(t, []string{"This"}, m.GetSuggestedReplacements())
	require.Equal(t, "This this", sent.GetText()[m.FromPos:m.ToPos])
	require.Equal(t, wordRepeatRuleENMessages()["repetition"], m.Message)
	require.Equal(t, wordRepeatRuleENMessages()["desc_repetition_short"], m.ShortMessage)
	// Category MISC + Duplication (Java WordRepeatRule ctor).
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSDuplication, rule.GetLocQualityIssueType())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
}
