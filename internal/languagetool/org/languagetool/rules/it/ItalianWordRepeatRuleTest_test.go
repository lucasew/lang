package it

// Twin of languagetool-language-modules/it/src/test/java/org/languagetool/rules/it/ItalianWordRepeatRuleTest.java
//
// Java setup:
//
//	Italian lang = new Italian();
//	ItalianWordRepeatRule rule = new ItalianWordRepeatRule(TestTools.getEnglishMessages(), lang);
//	JLanguageTool lt = new JLanguageTool(lang);
//	// rule.match(lt.getAnalyzedSentence(text))
//
// Tests must use real IT analysis via NewJLanguageTool("it").GetAnalyzedSentence, not AnalyzePlain.
import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// italianWordRepeatEnglishMessages ports TestTools.getEnglishMessages() keys used by
// ItalianWordRepeatRule / WordRepeatRule (MessagesBundle_en.properties).
func italianWordRepeatEnglishMessages() map[string]string {
	return map[string]string{
		"desc_repetition":       "Word repetition (e.g. 'will will')",
		"desc_repetition_short": "Word repetition",
		"repetition":            "Possible typo: you repeated a word.",
		"category_misc":         "Miscellaneous",
	}
}

// assertItalianWordRepeatMatches ports Java:
//
//	assertThat(rule.match(lt.getAnalyzedSentence(text)).length, is(expected))
func assertItalianWordRepeatMatches(t *testing.T, rule *ItalianWordRepeatRule, lt *languagetool.JLanguageTool, text string, expected int) {
	t.Helper()
	require.Equal(t, expected, len(rule.Match(lt.GetAnalyzedSentence(text))), "text=%q", text)
}

// TestItalianWordRepeatRule_Rule twins Java ItalianWordRepeatRuleTest.testRule (all 5 cases).
func TestItalianWordRepeatRule_Rule(t *testing.T) {
	// Java: new ItalianWordRepeatRule(TestTools.getEnglishMessages(), lang)
	rule := NewItalianWordRepeatRule(italianWordRepeatEnglishMessages())
	// Java: new JLanguageTool(lang)
	lt := languagetool.NewJLanguageTool("it")

	// Java: 0-match (IT ignore arms + super.ignore "Duran")
	assertItalianWordRepeatMatches(t, rule, lt, "Mi è sembrato così così", 0)
	assertItalianWordRepeatMatches(t, rule, lt, "Duran Duran", 0)
	assertItalianWordRepeatMatches(t, rule, lt, "Devi mescolare piano piano", 0)
	assertItalianWordRepeatMatches(t, rule, lt, "Seguo passo passo", 0)
	// Java: 1-match real repeat
	assertItalianWordRepeatMatches(t, rule, lt, "Mi mi è sembrato così", 1)
}

// via arm from Java ItalianWordRepeatRule.ignore comment (not in *Test.java but Java-visible).
func TestItalianWordRepeatRule_ViaIgnorePhrase(t *testing.T) {
	rule := NewItalianWordRepeatRule(italianWordRepeatEnglishMessages())
	lt := languagetool.NewJLanguageTool("it")
	assertItalianWordRepeatMatches(t, rule, lt, "Il lessico si andava via via modificando.", 0)
}

// ignoreAtIT mirrors Java-style:
//
//	tokens = lt.getAnalyzedSentence(input).getTokensWithoutWhitespace()
//	return rule.ignore(tokens, position)
func ignoreAtIT(t *testing.T, rule *ItalianWordRepeatRule, lt *languagetool.JLanguageTool, input string, position int) bool {
	t.Helper()
	tokens := lt.GetAnalyzedSentence(input).GetTokensWithoutWhitespace()
	return rule.Ignore(tokens, position)
}

// ignore() IT arms + super.ignore name list after ExtraIgnore returns false.
// Positions are the second token of the repeated pair (tokensWithoutWhitespace).
func TestItalianWordRepeatRule_Ignore(t *testing.T) {
	rule := NewItalianWordRepeatRule(italianWordRepeatEnglishMessages())
	lt := languagetool.NewJLanguageTool("it")

	// wordRepetitionOf exact equals at the second of the repeated pair
	// SENT_START=0: Mi=1 è=2 sembrato=3 così=4 così=5
	require.True(t, ignoreAtIT(t, rule, lt, "Mi è sembrato così così", 5))
	// Seguo=1 passo=2 passo=3
	require.True(t, ignoreAtIT(t, rule, lt, "Seguo passo passo", 3))
	// Devi=1 mescolare=2 piano=3 piano=4
	require.True(t, ignoreAtIT(t, rule, lt, "Devi mescolare piano piano", 4))
	// Il=1 lessico=2 si=3 andava=4 via=5 via=6
	require.True(t, ignoreAtIT(t, rule, lt, "Il lessico si andava via via modificando.", 6))

	// Non-listed case-folded repeat is not ignored (Mi=1 mi=2)
	require.False(t, ignoreAtIT(t, rule, lt, "Mi mi è sembrato così", 2))

	// super.ignore names (Java return super.ignore after IT checks)
	require.True(t, ignoreAtIT(t, rule, lt, "Duran Duran", 2))
	require.True(t, ignoreAtIT(t, rule, lt, "Phi Phi Islands", 2))
}

// wordRepetitionOf uses Token.equals (case-sensitive), not EqualFold.
func TestItalianWordRepeatRule_WordRepetitionOfCaseSensitive(t *testing.T) {
	lt := languagetool.NewJLanguageTool("it")

	// exact "passo" pair
	tokens := lt.GetAnalyzedSentence("Seguo passo passo").GetTokensWithoutWhitespace()
	require.True(t, wordRepetitionOfIT("passo", tokens, 3))
	require.False(t, wordRepetitionOfIT("passo", tokens, 2))
	require.False(t, wordRepetitionOfIT("piano", tokens, 3))

	// "PASSO PASSO" is not wordRepetitionOf("passo")
	tokensUpper := lt.GetAnalyzedSentence("PASSO PASSO").GetTokensWithoutWhitespace()
	require.False(t, wordRepetitionOfIT("passo", tokensUpper, 2))
	require.True(t, wordRepetitionOfIT("PASSO", tokensUpper, 2))

	// "Così così" is not wordRepetitionOf("così") (first token case differs)
	tokensMixed := lt.GetAnalyzedSentence("Così così").GetTokensWithoutWhitespace()
	require.False(t, wordRepetitionOfIT("così", tokensMixed, 2))
	require.False(t, wordRepetitionOfIT("Così", tokensMixed, 2))

	// Latin super list is case-sensitive via base wordRepetitionOf
	rule := NewItalianWordRepeatRule(nil)
	require.True(t, ignoreAtIT(t, rule, lt, "Duran Duran", 2))
	require.False(t, ignoreAtIT(t, rule, lt, "duran duran", 2))
	require.False(t, ignoreAtIT(t, rule, lt, "DURAN DURAN", 2))
}

// createRuleMatch outcomes: id, message, short message, span, suggestion.
func TestItalianWordRepeatRule_MatchSpanAndSuggestion(t *testing.T) {
	msgs := italianWordRepeatEnglishMessages()
	rule := NewItalianWordRepeatRule(msgs)
	lt := languagetool.NewJLanguageTool("it")
	text := "Mi mi è sembrato così"
	sent := lt.GetAnalyzedSentence(text)
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	m := matches[0]
	require.Equal(t, "ITALIAN_WORD_REPEAT_RULE", m.GetSpecificRuleId())
	require.Equal(t, msgs["repetition"], m.Message)
	require.Equal(t, msgs["desc_repetition_short"], m.ShortMessage)
	// Java: setSuggestedReplacement(prevToken) — surface of first "Mi"
	require.Equal(t, []string{"Mi"}, m.GetSuggestedReplacements())
	// Java: fromPos of first Mi through pos+prevToken.length() of second (UTF-16)
	require.Equal(t, "Mi mi", utf16SubstrIT(text, m.FromPos, m.ToPos))
	require.Equal(t, rule.WordRepeatRule, m.GetRule())
}

// utf16SubstrIT slices text by Java UTF-16 code-unit offsets (token startPos/toPos).
func utf16SubstrIT(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

// Java getId + MISC category + Duplication + estimateContextForSureMatch + description key.
// ItalianWordRepeatRule does not call addExamplePair.
func TestItalianWordRepeatRule_Metadata(t *testing.T) {
	msgs := italianWordRepeatEnglishMessages()
	rule := NewItalianWordRepeatRule(msgs)
	require.Equal(t, "ITALIAN_WORD_REPEAT_RULE", rule.GetID())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
	require.Equal(t, msgs["desc_repetition"], rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.CategoryMisc))
	require.Equal(t, "Miscellaneous", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSDuplication, rule.GetLocQualityIssueType())
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
}
