package ga

// Outcome twin of org.languagetool.rules.ga.IrishReplaceRule
// (no dedicated Java *RuleTest class; bar matches FA SimpleReplaceRule /
// PT EnglishContractionSpellingRule ACCEPT: metadata + official replace.txt
// outcomes + exact getMessage text).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// Official /ga/replace.txt keys are surface forms (e.g. bhúr, abalta).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official replace.txt rows.
func TestIrishReplaceRule_Rule(t *testing.T) {
	// Java: new IrishReplaceRule(messages, language)
	rule := NewIrishReplaceRule(nil)

	// correct sentences (fixed forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("ar bhur gcuid cainte."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tá sé ábalta."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair replace."))))

	// Java example: bhúr → bhur
	checkIrishReplace(t, rule, "ar bhúr gcuid cainte.", "bhur")
	// full Java example sentence
	checkIrishReplace(t, rule,
		"Níl beann agam oraibh, ar bhúr gcuid cainte.",
		"bhur")

	// Official replace.txt rows
	checkIrishReplace(t, rule, "Tá sé abalta.", "ábalta")
	checkIrishReplace(t, rule, "an aghaigh sin", "aghaidh")
	checkIrishReplace(t, rule, "tiocfaidh amarach", "amárach")

	// multi-suggestion key: amhainn=abhainn|amháin
	matchesMulti := rule.Match(languagetool.AnalyzePlain("an amhainn mór"))
	require.Len(t, matchesMulti, 1)
	require.Equal(t, []string{"abhainn", "amháin"}, matchesMulti[0].GetSuggestedReplacements())
	require.Equal(t,
		"Litriú mícheart: \"abhainn, amháin\".",
		matchesMulti[0].GetMessage())
	require.Equal(t, "Litriú mícheart", matchesMulti[0].GetShortMessage())

	// Java-visible createRuleMatch message/short for example-pair hit
	matches := rule.Match(languagetool.AnalyzePlain("ar bhúr gcuid cainte."))
	require.Len(t, matches, 1)
	require.Equal(t, "Litriú mícheart", matches[0].GetShortMessage())
	require.Equal(t, "Litriú mícheart: \"bhur\".", matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, IrishReplaceRuleID, base.GetID())
}

// checkIrishReplace asserts one match with the given first replacement and exact message.
func checkIrishReplace(t *testing.T, rule *IrishReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	require.Equal(t,
		"Litriú mícheart: \""+word+"\".",
		matches[0].GetMessage(),
		"message for %q", sentence)
	require.Equal(t, "Litriú mícheart", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, IrishReplaceRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of IrishReplaceRule constructor + overrides.
func TestIrishReplaceRule_Metadata(t *testing.T) {
	rule := NewIrishReplaceRule(nil)

	require.Equal(t, IrishReplaceRuleID, rule.GetID())
	require.Equal(t, "GA_REPLACE", rule.GetID())
	require.Equal(t, "Litriú mícheart, m.sh., \"agsu\" in áit \"agus\"", rule.GetDescription())
	require.Equal(t, "Litriú mícheart", rule.GetShort())
	// MessageFn with empty replacements (static surface for metadata probes)
	require.Equal(t, "Litriú mícheart: \"\".", rule.GetMessage())

	// Java getLocale() → Locale("GA")
	require.Equal(t, "GA", rule.GetLocaleCode())
	require.Equal(t, "GA", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords)
	require.Equal(t, "", rule.GetURL())

	// Java: Categories.TYPOS + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /ga/replace.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"bhur"}, ww["bhúr"])
	require.Equal(t, []string{"ábalta"}, ww["abalta"])
	require.Equal(t, []string{"aghaidh"}, ww["aghaigh"])
	require.Equal(t, []string{"amárach"}, ww["amarach"])
	require.Equal(t, []string{"abhainn", "amháin"}, ww["amhainn"])
	require.Equal(t, IrishReplaceFilePath, "/ga/replace.txt")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewIrishReplaceRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Litriú mícheart, m.sh., \"agsu\" in áit \"agus\"", ruleMsgs.GetDescription())
	require.Equal(t, "Litriú mícheart", ruleMsgs.GetShort())

	// Java addExamplePair(bhúr → bhur)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Níl beann agam oraibh, ar <marker>bhúr</marker> gcuid cainte.", inc[0].GetExample())
	require.Equal(t, []string{"bhur"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "Níl beann agam oraibh, ar <marker>bhur</marker> gcuid cainte.", corr[0].GetExample())
}

// Load does not panic; resource parses.
func TestIrishReplaceRule_Loads(t *testing.T) {
	rule := NewIrishReplaceRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
}
