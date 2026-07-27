package ga

// Outcome twin of org.languagetool.rules.ga.IrishFGBEqReplaceRule
// (no dedicated Java *RuleTest class; bar matches IrishReplaceRule ACCEPT:
// metadata + official replace-fgb-eq.txt outcomes + exact getMessage text).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// Official /ga/replace-fgb-eq.txt keys are surface forms (e.g. urlamh, urlaicean).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official replace-fgb-eq.txt rows.
func TestIrishFGBEqReplaceRule_Rule(t *testing.T) {
	// Java: new IrishFGBEqReplaceRule(messages, language)
	rule := NewIrishFGBEqReplaceRule(nil)

	// correct sentences (fixed forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("An bhfuil tú ullamh?"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("an t-urlacan sin"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair fgb-eq."))))

	// Java example: urlamh → ullamh
	checkIrishFGBEqReplace(t, rule, "An bhfuil tú urlamh?", "ullamh")
	// full Java example sentence surface
	checkIrishFGBEqReplace(t, rule, "An bhfuil tú urlamh?", "ullamh")

	// Official replace-fgb-eq.txt rows (all keys; file is small)
	checkIrishFGBEqReplace(t, rule, "hurlaicean anseo", "hurlacan")
	checkIrishFGBEqReplace(t, rule, "n-urlaicean anseo", "n-urlacan")
	checkIrishFGBEqReplace(t, rule, "t-urlaicean anseo", "t-urlacan")
	checkIrishFGBEqReplace(t, rule, "urlaicean anseo", "urlacan")
	checkIrishFGBEqReplace(t, rule, "hurlamh anseo", "hullamh")
	checkIrishFGBEqReplace(t, rule, "urlamha anseo", "ullamha")

	// Java-visible createRuleMatch message/short for example-pair hit
	matches := rule.Match(languagetool.AnalyzePlain("An bhfuil tú urlamh?"))
	require.Len(t, matches, 1)
	require.Equal(t, "Neamhchoitianta", matches[0].GetShortMessage())
	require.Equal(t, "Focal ceart ach tá \"ullamh\" níos coitianta.", matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, IrishFGBEqReplaceRuleID, base.GetID())
}

// checkIrishFGBEqReplace asserts one match with the given first replacement and exact message.
func checkIrishFGBEqReplace(t *testing.T, rule *IrishFGBEqReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	// Java: "Focal ceart ach tá \"" + join + "\" níos coitianta."
	require.Equal(t,
		"Focal ceart ach tá \""+word+"\" níos coitianta.",
		matches[0].GetMessage(),
		"message for %q", sentence)
	require.Equal(t, "Neamhchoitianta", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, IrishFGBEqReplaceRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of IrishFGBEqReplaceRule constructor + overrides.
func TestIrishFGBEqReplaceRule_Metadata(t *testing.T) {
	rule := NewIrishFGBEqReplaceRule(nil)

	require.Equal(t, IrishFGBEqReplaceRuleID, rule.GetID())
	require.Equal(t, "GA_FGB_EQ_REPLACE", rule.GetID())
	require.Equal(t, "Ceannfhocal FGB neamhchoitianta, m.sh., \"urlamh\" in áit \"ullamh\"", rule.GetDescription())
	require.Equal(t, "Neamhchoitianta", rule.GetShort())
	// MessageFn with empty replacements (static surface for metadata probes)
	require.Equal(t, "Focal ceart ach tá \"\" níos coitianta.", rule.GetMessage())

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

	// Official dictionary loaded (same keys as /ga/replace-fgb-eq.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"hurlacan"}, ww["hurlaicean"])
	require.Equal(t, []string{"n-urlacan"}, ww["n-urlaicean"])
	require.Equal(t, []string{"t-urlacan"}, ww["t-urlaicean"])
	require.Equal(t, []string{"urlacan"}, ww["urlaicean"])
	require.Equal(t, []string{"hullamh"}, ww["hurlamh"])
	require.Equal(t, []string{"ullamh"}, ww["urlamh"])
	require.Equal(t, []string{"ullamha"}, ww["urlamha"])
	require.Equal(t, 7, len(ww), "all replace-fgb-eq.txt keys must load")
	require.Equal(t, IrishFGBEqReplaceFilePath, "/ga/replace-fgb-eq.txt")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewIrishFGBEqReplaceRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Ceannfhocal FGB neamhchoitianta, m.sh., \"urlamh\" in áit \"ullamh\"", ruleMsgs.GetDescription())
	require.Equal(t, "Neamhchoitianta", ruleMsgs.GetShort())

	// Java addExamplePair(urlamh → ullamh)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "An bhfuil tú <marker>urlamh</marker>?", inc[0].GetExample())
	require.Equal(t, []string{"ullamh"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "An bhfuil tú <marker>ullamh</marker>?", corr[0].GetExample())
}

// Load does not panic; resource parses.
func TestIrishFGBEqReplaceRule_Loads(t *testing.T) {
	rule := NewIrishFGBEqReplaceRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
}
