package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWordRepeatRuleTest.java
//
// Java setup:
//   Language english = Languages.getLanguageForShortCode("en");
//   EnglishWordRepeatRule rule = new EnglishWordRepeatRule(TestTools.getEnglishMessages(), english);
//   JLanguageTool lt = new JLanguageTool(english);
//   assertGood/Bad: rule.match(lt.getAnalyzedSentence(sentence)).length is 0 or 1
//
// Go analysis path: tagging/en.AnalyzeEnglishSentence — EnglishWordTokenizer +
// EnglishTagger (english.dict) + hybrid disambiguator (same resources Java loads).
// POS-dependent ignore arms (her/had/that/can) rely on real tags, not surface invent.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

func engWordRepeatMessages() map[string]string {
	// Mirror TestTools English ResourceBundle keys used by WordRepeatRule.
	return map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition":       "Word repetition",
		"desc_repetition_short": "repetition",
		"category_misc":         "Miscellaneous",
	}
}

func engWordRepeatLT(t *testing.T) {
	t.Helper()
	// Ensure english.dict is discoverable (third_party/english-pos-dict or LANG_ENGLISH_DICT).
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree (third_party/english-pos-dict or inspiration)")
	}
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, tagen.EnglishTagWord(), "EnglishTagWord must load from english.dict")
}

// assertEWRMatches ports EnglishWordRepeatRuleTest.assertMatches:
// rule.match(lt.getAnalyzedSentence(sentence)).length == expectedMatches
func assertEWRMatches(t *testing.T, rule *EnglishWordRepeatRule, sentence string, expectedMatches int) {
	t.Helper()
	sent := tagen.AnalyzeEnglishSentence(sentence)
	require.NotNil(t, sent, "AnalyzeEnglishSentence(%q)", sentence)
	matches := rule.Match(sent)
	require.Equal(t, expectedMatches, len(matches), "assertMatches %q", sentence)
	// Java-visible createRuleMatch: suggested replacement is the previous token surface.
	if expectedMatches == 1 && len(matches) == 1 {
		require.NotEmpty(t, matches[0].GetSuggestedReplacements())
		// Span covers the repeated pair: prevPos .. pos+len(prevToken) (UTF-16).
		require.GreaterOrEqual(t, matches[0].ToPos, matches[0].FromPos)
		require.Equal(t, "ENGLISH_WORD_REPEAT_RULE", rule.GetID())
	}
}

func assertEWRGood(t *testing.T, rule *EnglishWordRepeatRule, sentence string) {
	t.Helper()
	assertEWRMatches(t, rule, sentence, 0)
}

func assertEWRBad(t *testing.T, rule *EnglishWordRepeatRule, sentence string) {
	t.Helper()
	assertEWRMatches(t, rule, sentence, 1)
}

func TestEnglishWordRepeatRule_RepeatRule(t *testing.T) {
	engWordRepeatLT(t)
	rule := NewEnglishWordRepeatRule(engWordRepeatMessages())

	// Order and strings match Java EnglishWordRepeatRuleTest.testRepeatRule.
	assertEWRGood(t, rule, "This is a test.")
	assertEWRGood(t, rule, "If I had had time, I would have gone to see him.")
	assertEWRGood(t, rule, "I don't think that that is a problem.")
	assertEWRGood(t, rule, "He also said that Azerbaijan had fulfilled a task he set, which was that that their defense budget should exceed the entire state budget of Armenia.")
	assertEWRGood(t, rule, "Just as if that was proof that that English was correct.")
	assertEWRGood(t, rule, "It was noticed after more than a month that that promise had not been carried out.")
	assertEWRGood(t, rule, "It was said that that lady was an actress.")
	assertEWRGood(t, rule, "Kurosawa's three consecutive movies after Seven Samurai had not managed to capture Japanese audiences in the way that that film had.")
	assertEWRGood(t, rule, "The can can hold the water.")
	assertEWRGood(t, rule, "May May awake up?")
	assertEWRGood(t, rule, "May may awake up.")
	assertEWRGood(t, rule, "The cat does meow meow")
	assertEWRGood(t, rule, "Hah Hah")
	assertEWRGood(t, rule, "Hip Hip Hooray")
	assertEWRBad(t, rule, "Hip Hip")
	assertEWRGood(t, rule, "It's S.T.E.A.M.")
	assertEWRGood(t, rule, "Ok ok ok!")
	assertEWRGood(t, rule, "O O O")
	assertEWRGood(t, rule, "Alice and Bob had had a long-standing relationship.")
	assertEWRBad(t, rule, "I may may awake up.")
	assertEWRBad(t, rule, "That is May May.")
	assertEWRGood(t, rule, "Will Will awake up?")
	assertEWRGood(t, rule, "Will will awake up.")
	assertEWRBad(t, rule, "I will will awake up.")
	assertEWRBad(t, rule, "Please wait wait for me.")
	assertEWRGood(t, rule, "Wait wait!")
	assertEWRBad(t, rule, "That is Will Will.")
	assertEWRBad(t, rule, "I will will hold the ladder.")
	assertEWRBad(t, rule, "You can feel confident that that this administration will continue to support a free and open Internet.")
	assertEWRBad(t, rule, "This is is a test.")
	assertEWRGood(t, rule, "b a s i c a l l y")
	assertEWRGood(t, rule, "You can contact E.ON on Instagram.")
	assertEWRBad(t, rule, "But I i was not sure.")
	assertEWRBad(t, rule, "I I am the best.")
	assertEWRGood(t, rule, "In a land far far away.")
	assertEWRGood(t, rule, "I love you so so much.")
	assertEWRGood(t, rule, "Aye aye, sir!")
	assertEWRGood(t, rule, "What Tom did didn't seem to bother Mary at all.")
	assertEWRGood(t, rule, "Whatever you do don't leave the lid up on the toilet!")
	assertEWRGood(t, rule, "Keep your chin up and whatever you do don't doubt yourself or your actions.")
	assertEWRGood(t, rule, "I know that that can't really happen.")
	assertEWRGood(t, rule, "Please pass her her phone.")
	assertEWRGood(t, rule, "I have visited Bora Bora.")
}

// Java createRuleMatch outcomes for a clear false-alarm: positions + suggestion.
func TestEnglishWordRepeatRule_MatchSpanAndSuggestion(t *testing.T) {
	engWordRepeatLT(t)
	rule := NewEnglishWordRepeatRule(engWordRepeatMessages())
	sent := tagen.AnalyzeEnglishSentence("This is is a test.")
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	m := matches[0]
	// "is is" — from start of first is to end of second is (UTF-16 length of prevToken).
	require.Equal(t, []string{"is"}, m.GetSuggestedReplacements())
	require.Equal(t, "is is", sent.GetText()[m.FromPos:m.ToPos])
	require.Equal(t, engWordRepeatMessages()["repetition"], m.Message)
	require.Equal(t, engWordRepeatMessages()["desc_repetition_short"], m.ShortMessage)
}

// Without POS tags, her/had/that/can arms fail closed (no invent). Uses EN tokenizer
// so surfaces match Java; tags absent so posIsIn is false.
func TestEnglishWordRepeatRule_PosIsInFailClosed(t *testing.T) {
	rule := NewEnglishWordRepeatRule(map[string]string{"repetition": "rep"})
	analyze := func(s string) *languagetool.AnalyzedSentence {
		return languagetool.AnalyzeWithTokenizer(s, entok.NewEnglishWordTokenizer())
	}
	require.Equal(t, 1, len(rule.Match(analyze("Please pass her her phone."))))
	require.Equal(t, 1, len(rule.Match(analyze("If I had had time."))))
	require.Equal(t, 1, len(rule.Match(analyze("The can can hold water."))))
	// n't ignore is surface-only (no POS) — must still suppress did/did + n't.
	require.Equal(t, 0, len(rule.Match(analyze("What Tom did didn't seem to bother Mary at all."))))
}

// Java EnglishWordRepeatRule: ENGLISH_WORD_REPEAT_RULE id + is is example pair.
func TestEnglishWordRepeatRule_Metadata(t *testing.T) {
	rule := NewEnglishWordRepeatRule(nil)
	require.Equal(t, "ENGLISH_WORD_REPEAT_RULE", rule.GetID())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "This <marker>is is</marker> just an example sentence.", inc[0].GetExample())
	require.Equal(t, []string{"is"}, inc[0].GetCorrections())
	require.Equal(t, "This <marker>is</marker> just an example sentence.", rule.GetCorrectExamples()[0].GetExample())
	// Category MISC + Duplication (Java WordRepeatRule ctor via super).
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSDuplication, rule.GetLocQualityIssueType())
}
