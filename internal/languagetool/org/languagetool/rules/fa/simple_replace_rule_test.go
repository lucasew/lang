package fa

// Outcome twin of org.languagetool.rules.fa.SimpleReplaceRule
// (no dedicated Java *RuleTest class; bar matches other resource-backed leaves
// without *RuleTest: Java example pair + official replace.txt outcomes + metadata).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; POS/lemmas only matter when checkLemmas finds
// lemma keys — official /fa/replace.txt keys are surface forms (e.g. حاظر).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official replace.txt rows.
func TestSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceRule(messages, language)
	rule := NewSimpleReplaceRule(nil)

	// correct sentence (fixed form from example pair / unrelated) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("وی حاضر به همکاری شد."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("جمله بدون غلط املائی."))))

	// Java example: حاظر → حاضر
	checkFASimpleReplace(t, rule, "وی حاظر به همکاری شد.", "حاضر")

	// Official replace.txt rows
	checkFASimpleReplace(t, rule, "این آزوقه تازه است.", "آذوقه")
	checkFASimpleReplace(t, rule, "اقای مدیر آمد.", "آقای")
	checkFASimpleReplace(t, rule, "جستوجو در وب.", "جستجو")
	checkFASimpleReplace(t, rule, "حظور در جلسه.", "حضور")

	// Java-visible createRuleMatch message/short for example-pair hit
	matches := rule.Match(languagetool.AnalyzePlain("وی حاظر به همکاری شد."))
	require.Len(t, matches, 1)
	require.Equal(t, "اشتباه محتمل املائی", matches[0].GetShortMessage())
	require.Equal(t, "اشتباه محتمل املائی پیداشده: حاضر.", matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, FASimpleReplaceID, base.GetID())
}

// checkFASimpleReplace asserts one match with the given first replacement.
func checkFASimpleReplace(t *testing.T, rule *SimpleReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	require.Equal(t, FASimpleReplaceID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of SimpleReplaceRule constructor + overrides.
func TestSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)

	require.Equal(t, FASimpleReplaceID, rule.GetID())
	require.Equal(t, "FA_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "اشتباه محتمل املائی", rule.GetDescription())
	require.Equal(t, "اشتباه محتمل املائی", rule.GetShort())
	// MessageFn with empty replacements (static surface for metadata probes)
	require.Equal(t, "اشتباه محتمل املائی پیداشده: .", rule.GetMessage())

	// Java getLocale() → Locale.getDefault() (not Locale("fa")); LanguageCode empty.
	require.Equal(t, "", rule.GetLocaleCode())
	require.Equal(t, "", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.True(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords)
	require.Equal(t, "", rule.GetURL())

	// Java: Categories.CONFUSED_WORDS + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CONFUSED_WORDS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /fa/replace.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"حاضر"}, ww["حاظر"])
	require.Equal(t, []string{"آذوقه"}, ww["آزوقه"])
	require.Equal(t, []string{"آقای"}, ww["اقای"])
	require.Equal(t, []string{"حضور"}, ww["حظور"])

	// Category name honors supplied ResourceBundle map
	faMsgs := map[string]string{
		"category_confused_words": "کلمات گیج‌کننده",
	}
	ruleMsgs := NewSimpleReplaceRule(faMsgs)
	require.Equal(t, "CONFUSED_WORDS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "کلمات گیج‌کننده", ruleMsgs.GetCategory().GetName())

	// Java addExamplePair
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, []string{"حاضر"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
}
