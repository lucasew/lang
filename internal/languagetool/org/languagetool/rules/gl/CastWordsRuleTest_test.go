package gl

// Outcome twin of org.languagetool.rules.gl.CastWordsRule
// (no dedicated Java *RuleTest class; bar matches other resource-backed leaves
// without *RuleTest: official /gl/spanish.txt outcomes + metadata from the rule class).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; POS/lemmas only matter when checkLemmas finds
// lemma keys — official /gl/spanish.txt keys are surface forms (e.g. acera).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official spanish.txt rows (Java sample acera=beirarrúa).
func TestCastWordsRule_Rule(t *testing.T) {
	// Java: new CastWordsRule(messages, language)
	rule := NewCastWordsRule(nil)

	// correct sentence (fixed form / unrelated) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Camiñaba pola beirarrúa."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A oliva é boa."))))

	// Official spanish.txt: acera=beirarrúa (prior smoke sample / Java-visible row)
	checkCastWords(t, rule, "Camiñaba pola acera.", "beirarrúa")

	// Additional official rows for outcome fidelity
	checkCastWords(t, rule, "Comprou unha aceituna.", "oliva")
	checkCastWords(t, rule, "O alambre está roto.", "arame")
	checkCastWords(t, rule, "O altavoz non funciona.", "altofalante")

	// Java-visible createRuleMatch message/short/id/category for acera hit
	matches := rule.Match(languagetool.AnalyzePlain("Camiñaba pola acera."))
	require.Len(t, matches, 1)
	require.Equal(t, "Castelanismos léxicos", matches[0].GetShortMessage())
	require.Equal(t,
		"'acera' é un castelanismo. Empregue no seu sitio: beirarrúa.",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, CastWordsRuleID, base.GetID())
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "MISC", base.GetCategory().GetID().String())
}

// checkCastWords asserts one match with the given first replacement.
func checkCastWords(t *testing.T, rule *CastWordsRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	require.Equal(t, CastWordsRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of CastWordsRule constructor + overrides.
func TestCastWordsRule_Metadata(t *testing.T) {
	rule := NewCastWordsRule(nil)

	require.Equal(t, CastWordsRuleID, rule.GetID())
	require.Equal(t, "GL_CAST_WORDS", rule.GetID())
	require.Equal(t, "Corrección de erros léxicos (castelanismos).", rule.GetDescription())
	require.Equal(t, "Castelanismos léxicos", rule.GetShort())
	// MessageFn static surface (empty token/replacements)
	require.Equal(t, "'' é un castelanismo. Empregue no seu sitio: .", rule.GetMessage())
	// Sample replacements (Java getMessage with real args)
	require.Equal(t,
		"'acera' é un castelanismo. Empregue no seu sitio: beirarrúa.",
		rule.MessageFn("acera", []string{"beirarrúa"}))
	require.Equal(t,
		"'x' é un castelanismo. Empregue no seu sitio: a, b.",
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

	// Official dictionary loaded (same keys as /gl/spanish.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"beirarrúa"}, ww["acera"])
	require.Equal(t, []string{"oliva"}, ww["aceituna"])
	require.Equal(t, []string{"arame"}, ww["alambre"])
	require.Equal(t, []string{"altofalante"}, ww["altavoz"])

	// Category name honors supplied ResourceBundle map
	glMsgs := map[string]string{
		"category_misc": "Miscelánea",
	}
	ruleMsgs := NewCastWordsRule(glMsgs)
	require.Equal(t, "MISC", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Miscelánea", ruleMsgs.GetCategory().GetName())

	// Java does not call addExamplePair
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())

	// Path constant twin of loadFromPath("/gl/spanish.txt")
	require.Equal(t, "/gl/spanish.txt", CastWordsFilePath)
}
