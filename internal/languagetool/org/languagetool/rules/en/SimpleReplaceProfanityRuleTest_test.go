package en

// Outcome twin of org.languagetool.rules.en.SimpleReplaceProfanityRule
// (no dedicated Java unit test; fields + AbstractSimpleReplaceRule2 match outcomes
// mirror SimpleReplaceRuleTest / EnglishDiacriticsRuleTest style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(en).getAnalyzedSentence for this rule’s surface matching.
// SimpleReplaceProfanityRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official
// replace_profanity.txt multiword keys (e.g. "basket case") tokenize the same under
// WordTokenizer as EN. Phrases are key-only (no '='); setRuleHasSuggestions(false).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official replace_profanity.txt phrase-only rows.
func TestSimpleReplaceProfanityRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceProfanityRule(TestTools.getMessages("en"), new English())
	//       super(messages, language); //useSubRuleSpecificIds(); setRuleHasSuggestions(false); …
	rule := NewSimpleReplaceProfanityRule(nil)

	// correct sentences (no listed phrase / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nothing offensive here."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Please discourage that behavior."))))

	// Official replace_profanity.txt key-only rows (no suggestions side).
	// Prefer relatively mild listed phrases still present in the official file.
	checkEnglishProfanityMatch(t, rule, "That albo was rude.")
	checkEnglishProfanityMatch(t, rule, "He is a basket case.")
	checkEnglishProfanityMatch(t, rule, "The word berdache appears in the list.")

	// Java-visible createMatch extras (no useSubRuleSpecificIds + fixed message/short)
	matches := rule.Match(languagetool.AnalyzePlain("That albo was rude."))
	require.Len(t, matches, 1)
	require.Equal(t, "Profanity", matches[0].GetShortMessage())
	require.Equal(t, "This expression can be considered offensive.", matches[0].GetMessage())
	require.Contains(t, matches[0].GetMessage(), "offensive")
	// Empty suggestions: setRuleHasSuggestions(false) skips setSuggestedReplacements
	require.Empty(t, matches[0].GetSuggestedReplacements())
	// No SpecificIdRule — useSubRuleSpecificIds is commented out in Java
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	// createMatch uses the AbstractSimpleReplaceRule2 receiver (id PROFANITY)
	require.Equal(t, EnglishProfanityRuleID, matches[0].GetSpecificRuleId())
	require.Equal(t, "PROFANITY", matches[0].GetSpecificRuleId())
	require.Equal(t, 5, matches[0].GetFromPos())
	require.Equal(t, 9, matches[0].GetToPos())

	// Multiword key-only ("basket case")
	matches = rule.Match(languagetool.AnalyzePlain("He is a basket case."))
	require.Len(t, matches, 1)
	require.Equal(t, "Profanity", matches[0].GetShortMessage())
	require.Equal(t, "This expression can be considered offensive.", matches[0].GetMessage())
	require.Empty(t, matches[0].GetSuggestedReplacements())
	require.Equal(t, EnglishProfanityRuleID, matches[0].GetSpecificRuleId())
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 19, matches[0].GetToPos())
}

// checkEnglishProfanityMatch asserts one match, offensive message, empty suggestions.
func checkEnglishProfanityMatch(t *testing.T, rule *SimpleReplaceProfanityRule, sentence string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Empty(t, matches[0].GetSuggestedReplacements(),
		"Expected empty suggestions for profanity sentence: %s", sentence)
	require.Contains(t, matches[0].GetMessage(), "offensive",
		"Expected offensive message for sentence: %s", sentence)
	require.Equal(t, "This expression can be considered offensive.", matches[0].GetMessage())
	require.Equal(t, "Profanity", matches[0].GetShortMessage())
}

// Field twin of SimpleReplaceProfanityRule constructor + overrides.
func TestSimpleReplaceProfanityRule_Metadata(t *testing.T) {
	rule := NewSimpleReplaceProfanityRule(nil)

	require.Equal(t, EnglishProfanityRuleID, rule.GetID())
	require.Equal(t, "PROFANITY", rule.GetID())
	require.Equal(t, "Profanity", rule.GetDescription())
	require.Equal(t, "Profanity", rule.GetShort())
	require.Equal(t, "This expression can be considered offensive.", rule.GetMessage())
	// Java getLocale() → new Locale("EN")
	require.Equal(t, "EN", rule.GetLocaleCode())
	require.Equal(t, []string{EnglishProfanityFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/en/replace_profanity.txt"}, rule.GetFileNames())
	// Java: //useSubRuleSpecificIds(); — left default false
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, EnglishProfanityURL, rule.GetURL())
	require.Equal(t,
		"https://en.wiktionary.org/wiki/Category:English_offensive_terms",
		rule.GetURL())
	require.Contains(t, rule.GetURL(), "English_offensive_terms")

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	// setTags(Collections.singletonList(Tag.picky))
	require.True(t, rule.HasTag(rules.TagPicky))
	require.Equal(t, []rules.Tag{rules.TagPicky}, rule.GetTags())

	// setRuleHasSuggestions(false)
	require.NotNil(t, rule.RuleHasSuggestions)
	require.False(t, *rule.RuleHasSuggestions)

	// Java SimpleReplaceProfanityRule ctor adds no example pairs
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
}
