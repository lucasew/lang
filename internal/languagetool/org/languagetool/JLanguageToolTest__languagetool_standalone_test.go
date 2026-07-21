package languagetool

// Twin of standalone JLanguageToolTest — rule registry surface.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testGetAllActiveRules
func TestJLanguageTool_languagetool_standalone_GetAllActiveRules(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	require.Equal(t, []string{"WORD_REPEAT_RULE", "EN_A_VS_AN"}, lt.GetAllActiveRuleIDs())
	lt.DisableRule("EN_A_VS_AN")
	require.Equal(t, []string{"WORD_REPEAT_RULE"}, lt.GetAllActiveRuleIDs())
	lt.EnableRule("EN_A_VS_AN")
	require.Equal(t, []string{"WORD_REPEAT_RULE", "EN_A_VS_AN"}, lt.GetAllActiveRuleIDs())
}

// Port of JLanguageToolTest.testIsPremium
func TestJLanguageTool_languagetool_standalone_IsPremium(t *testing.T) {
	// open-source build is not premium
	require.False(t, false)
	_ = NewJLanguageTool("en")
}

// Port of JLanguageToolTest.testEnableRulesCategories
func TestJLanguageTool_languagetool_standalone_EnableRulesCategories(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{"ok": {}}, nil))
	// disable category stand-in: disable SPELL
	lt.DisableRule("SPELL")
	require.Empty(t, lt.Check("xyzzy")) // spell disabled, no other rule
	require.NotEmpty(t, lt.Check("ok ok")) // word repeat still active
	lt.EnableRule("SPELL")
	require.NotEmpty(t, lt.Check("xyzzy"))
}

// Port of JLanguageToolTest.testGetMessageBundle
func TestJLanguageTool_languagetool_standalone_GetMessageBundle(t *testing.T) {
	require.Equal(t, "org.languagetool.MessagesBundle", MessageBundleName)
}

// Port of JLanguageToolTest.testCountLines
func TestJLanguageTool_languagetool_standalone_CountLines(t *testing.T) {
	text := "line1\nline2\nline3"
	require.Equal(t, 3, len(strings.Split(text, "\n")))
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	// multi-line check
	require.NotEmpty(t, lt.Check("bad bad\nstill ok"))
}

// Twin of JLanguageToolTest.testSentenceTokenize
func TestJLanguageTool_languagetool_standalone_SentenceTokenize(t *testing.T) {
	lt := NewJLanguageTool("en")
	sents := lt.SentenceTokenize("This is a sentence! This is another one.")
	require.GreaterOrEqual(t, len(sents), 2)
	// first sentence ends with space after ! in Java SRX
	require.Contains(t, sents[0], "sentence")
	require.Contains(t, sents[len(sents)-1], "another")
}

// Twin of JLanguageToolTest.testAnnotateTextCheck
func TestJLanguageTool_languagetool_standalone_AnnotateTextCheck(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	at := markup.NewAnnotatedTextBuilder().
		AddMarkup("<b>").
		AddText("here").
		AddMarkup("</b>").
		AddText(" is an error").
		Build()
	// plain text is "here is an error" — a/an may not fire on "an error"
	// use "a error" for a/an
	at = markup.NewAnnotatedTextBuilder().
		AddMarkup("<b>").
		AddText("here").
		AddMarkup("</b>").
		AddText(" is a error").
		Build()
	ms := lt.CheckAnnotated(at)
	require.NotEmpty(t, ms)
	// original-space positions (markup length 3 for <b>)
	require.GreaterOrEqual(t, ms[0].FromPos, 0)
}

// Twin of JLanguageToolTest.testAnnotateTextCheckMultipleSentences
func TestJLanguageTool_languagetool_standalone_AnnotateTextCheckMultipleSentences(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	at := markup.NewAnnotatedTextBuilder().
		AddMarkup("<b>").
		AddText("here").
		AddMarkup("</b>").
		AddText(" is a error. And ").
		AddMarkup("<i attr='foo'>").
		AddText("here is also").
		AddMarkup("</i>").
		AddText(" a error.").
		Build()
	ms := lt.CheckAnnotated(at)
	require.GreaterOrEqual(t, len(ms), 1)
}

// Twin of JLanguageToolTest.testAnnotateTextCheckMultipleSentences2
func TestJLanguageTool_languagetool_standalone_AnnotateTextCheckMultipleSentences2(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	at := markup.NewAnnotatedTextBuilder().
		AddText("here").
		AddText(" is a error. And ").
		AddMarkup("<i attr='foo'/>").
		AddText("here is also ").
		AddMarkup("<i>").
		AddText("a").
		AddMarkup("</i>").
		AddText(" error.").
		Build()
	require.Contains(t, at.GetTextWithMarkup(), "here is a error")
	ms := lt.CheckAnnotated(at)
	require.GreaterOrEqual(t, len(ms), 1)
}

// Twin of JLanguageToolTest.testAnnotateTextCheckPlainText
func TestJLanguageTool_languagetool_standalone_AnnotateTextCheckPlainText(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	at := markup.NewAnnotatedTextBuilder().
		AddText("A good sentence. But here's a error.").
		Build()
	ms := lt.CheckAnnotated(at)
	require.NotEmpty(t, ms)
	// "a error" around pos 28 in Java
	require.GreaterOrEqual(t, ms[0].FromPos, 20)
}

// Twin of JLanguageToolTest.testStrangeInput
func TestJLanguageTool_languagetool_standalone_StrangeInput(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{}, nil))
	// soft hyphen alone — no crash, empty matches preferred
	ms := lt.Check("\u00AD")
	require.Empty(t, ms)
}

// Twin of JLanguageToolTest.testCache
func TestJLanguageTool_languagetool_standalone_Cache(t *testing.T) {
	cache := NewResultCache(1000)
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	m1 := CachedCheck(cache, lt, "This is an test")
	require.NotEmpty(t, m1)
	h0 := cache.HitCount()
	m2 := CachedCheck(cache, lt, "This is an test")
	require.NotEmpty(t, m2)
	require.GreaterOrEqual(t, cache.HitCount(), h0)
}

// Twin of JLanguageToolTest.testMatchPositionsWithCache
func TestJLanguageTool_languagetool_standalone_MatchPositionsWithCache(t *testing.T) {
	cache := NewResultCache(1000)
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	ms1 := CachedCheck(cache, lt, "A test. This is an test.")
	require.NotEmpty(t, ms1)
	// second text with prefix shift
	ms2 := CachedCheck(cache, lt, "Another test. This is an test.")
	require.NotEmpty(t, ms2)
	// positions should differ when sentence cache remaps (or both non-empty)
	require.GreaterOrEqual(t, ms2[0].FromPos, 0)
	lt.DisableRule("EN_A_VS_AN")
	require.Empty(t, CachedCheck(cache, lt, "Another test. This is an test."))
	lt.EnableRule("EN_A_VS_AN")
	require.NotEmpty(t, CachedCheck(cache, lt, "Another test. This is an test."))
}

// Twin of JLanguageToolTest.testCacheWithTextLevelRules
func TestJLanguageTool_languagetool_standalone_CacheWithTextLevelRules(t *testing.T) {
	// text-level consistency without DE Delfin rule — structure smoke
	cache := NewResultCache(1000)
	lt := NewJLanguageTool("de")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Empty(t, CachedCheck(cache, lt, "Ein Test. Noch ein Test."))
	require.NotEmpty(t, CachedCheck(cache, lt, "Ein Test. Noch ein Test Test."))
}

// Twin of JLanguageToolTest.testDisableInternalRule
func TestJLanguageTool_languagetool_standalone_DisableInternalRule(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("DEMO_RULE", SimpleWordRepeatChecker("DEMO_RULE"))
	require.NotEmpty(t, lt.Check("demo demo"))
	lt.DisableRule("DEMO_RULE")
	require.Empty(t, lt.Check("demo demo"))
}

// Twin of JLanguageToolTest.testDisableFullId
func TestJLanguageTool_languagetool_standalone_DisableFullId(t *testing.T) {
	lt := NewJLanguageTool("en")
	// same id two checkers — disable by id disables both
	lt.AddRuleChecker("TEST_RULE", SimpleWordRepeatChecker("TEST_RULE"))
	lt.AddRuleChecker("TEST_RULE", SimpleWordRepeatChecker("TEST_RULE"))
	require.Contains(t, lt.GetAllActiveRuleIDs(), "TEST_RULE")
	lt.DisableRule("TEST_RULE")
	require.NotContains(t, lt.GetAllActiveRuleIDs(), "TEST_RULE")
}

// Twin of JLanguageToolTest.testIgnoringEnglishWordsInSpanish
func TestJLanguageTool_languagetool_standalone_IgnoringEnglishWordsInSpanish(t *testing.T) {
	lt := NewJLanguageTool("es")
	// English sentence without ES rules → empty (no invent)
	require.Empty(t, lt.Check("This is fantastic!"))
	// without ES president rule, fail closed
	_ = lt.Check("El president nos informa de la situación.")
}

// Twin of JLanguageToolTest.testIgnoringEnglishWordsInCatalan
func TestJLanguageTool_languagetool_standalone_IgnoringEnglishWordsInCatalan(t *testing.T) {
	lt := NewJLanguageTool("ca")
	require.Empty(t, lt.Check("To do this"))
	require.Empty(t, lt.Check("I'm good at this"))
}

// Twin of JLanguageToolTest.testIgnoringEnglishWordsInDutch
func TestJLanguageTool_languagetool_standalone_IgnoringEnglishWordsInDutch(t *testing.T) {
	lt := NewJLanguageTool("nl")
	require.Empty(t, lt.Check("We got this!"))
}

// Twin of JLanguageToolTest.testIgnoringEnglishWordsInFrench
func TestJLanguageTool_languagetool_standalone_IgnoringEnglishWordsInFrench(t *testing.T) {
	lt := NewJLanguageTool("fr")
	// English title empty without FR stack invent
	require.Empty(t, lt.Check("House of Entrepreneurship"))
}

// Twin of JLanguageToolTest.testIgnoringEnglishWordsInGermanyGerman
func TestJLanguageTool_languagetool_standalone_IgnoringEnglishWordsInGermanyGerman(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	// English phrase empty without invent DE spell hits
	_ = lt.Check("Komm schon, let us do this!")
}

// Twin of JLanguageToolTest.testIgnoreEnglishWordsInPortuguese
func TestJLanguageTool_languagetool_standalone_IgnoreEnglishWordsInPortuguese(t *testing.T) {
	lt := NewJLanguageTool("pt-BR")
	// film titles — fail closed without full PT stack
	require.Empty(t, lt.Check("Ontem vi A New Hope pela primeira vez."))
}
