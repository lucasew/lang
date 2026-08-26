package gl

// Outcome twin of org.languagetool.rules.gl.SimpleReplaceRule
// (no dedicated Java *RuleTest class; bar matches other resource-backed leaves
// without *RuleTest: official /gl/words.txt outcomes + metadata from the rule class).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; POS/lemmas only matter when checkLemmas finds
// lemma keys — official /gl/words.txt keys are surface forms (e.g. alomenos).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official words.txt rows (Java sample alomenos=polo menos).
func TestSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceRule(messages, language)
	rule := NewSimpleReplaceRule(nil)

	// correct sentence (fixed form / unrelated) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Hai polo menos tres persoas."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("O ambiente é bo."))))

	// Official words.txt: alomenos=polo menos (prior smoke sample / Java-visible row)
	checkSimpleReplace(t, rule, "Hai alomenos tres persoas.", "polo menos")

	// Additional official rows for outcome fidelity
	checkSimpleReplace(t, rule, "Foi un achádego importante.", "achado")
	checkSimpleReplace(t, rule, "O ambente é bo.", "ambiente")
	checkSimpleReplace(t, rule, "Saíu anque chovía.", "aínda que")

	// Capitalized surface: ASR applies case of token to suggestion ("Polo menos")
	checkSimpleReplace(t, rule, "Alomenos tres persoas.", "Polo menos")

	// Java-visible createRuleMatch message/short/id/category for alomenos hit
	matches := rule.Match(languagetool.AnalyzePlain("Hai alomenos tres persoas."))
	require.Len(t, matches, 1)
	require.Equal(t, "Erros léxicos", matches[0].GetShortMessage())
	require.Equal(t,
		"'alomenos' non existe en galego. Talvez quería vostede dicir: polo menos.",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, SimpleReplaceRuleID, base.GetID())
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "MISC", base.GetCategory().GetID().String())
}

// checkSimpleReplace asserts one match with the given first replacement.
func checkSimpleReplace(t *testing.T, rule *SimpleReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	require.Equal(t, SimpleReplaceRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of SimpleReplaceRule constructor + overrides.
func TestSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)

	require.Equal(t, SimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "GL_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Corrección de erros léxicos (barbarismos).", rule.GetDescription())
	require.Equal(t, "Erros léxicos", rule.GetShort())
	// MessageFn static surface (empty token/replacements)
	require.Equal(t, "'' non existe en galego. Talvez quería vostede dicir: .", rule.GetMessage())
	// Sample replacements (Java getMessage with real args)
	require.Equal(t,
		"'alomenos' non existe en galego. Talvez quería vostede dicir: polo menos.",
		rule.MessageFn("alomenos", []string{"polo menos"}))
	require.Equal(t,
		"'x' non existe en galego. Talvez quería vostede dicir: a, b.",
		rule.MessageFn("x", []string{"a", "b"}))

	// Java getLocale() → Locale("gl")
	require.Equal(t, "gl", rule.GetLocaleCode())
	require.Equal(t, "gl", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.True(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords)
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// Java relies on AbstractSimpleReplaceRule super: Categories.MISC; no ITS override
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /gl/words.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"polo menos"}, ww["alomenos"])
	require.Equal(t, []string{"achado"}, ww["achádego"])
	require.Equal(t, []string{"ambiente"}, ww["ambente"])
	require.Equal(t, []string{"aínda que"}, ww["anque"])

	// Category name honors supplied ResourceBundle map
	glMsgs := map[string]string{
		"category_misc": "Miscelánea",
	}
	ruleMsgs := NewSimpleReplaceRule(glMsgs)
	require.Equal(t, "MISC", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Miscelánea", ruleMsgs.GetCategory().GetName())

	// Java does not call addExamplePair
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())

	// Path constant twin of loadFromPath("/gl/words.txt")
	require.Equal(t, "/gl/words.txt", SimpleReplaceFilePath)
}
