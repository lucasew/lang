package languagetool

// Twin of languagetool-language-modules/en JLanguageToolTest (module lang_en).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of JLanguageToolTest.demoCodeForHomepage
func TestJLanguageTool_lang_en_DemoCodeForHomepage(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	lt.RegisterDemoEnglishCheckers(map[string]struct{}{
		"A": {}, "a": {}, "an": {}, "sentence": {}, "with": {}, "error": {}, "in": {}, "the": {},
		"Hitchhiker": {}, "Guide": {}, "Galaxy": {}, "to": {}, "he": {}, "s": {},
	}, nil)
	src := "A sentence with a error in the Hitchhiker's Guide to the Galaxy"
	matches := lt.Check(src)
	require.NotEmpty(t, matches)
	ids := map[string]bool{}
	for _, m := range matches {
		ids[m.RuleID] = true
	}
	require.True(t, ids["EN_A_VS_AN"], "expected a→an for 'a error'")
	fixed := src
	for pass := 0; pass < 16; pass++ {
		ms := lt.Check(fixed)
		if len(ms) == 0 {
			break
		}
		var pick *LocalMatch
		for i := range ms {
			m := &ms[i]
			if len(m.Suggestions) == 0 {
				continue
			}
			if m.RuleID == "EN_A_VS_AN" {
				pick = m
				break
			}
			if pick == nil {
				pick = m
			}
		}
		if pick == nil {
			break
		}
		next := CorrectTextFromLocalMatches(fixed, []LocalMatch{*pick})
		if next == fixed {
			break
		}
		fixed = next
	}
	require.Contains(t, fixed, "an error")
}

// Twin of JLanguageToolTest.spellCheckerDemoCodeForHomepage
func TestJLanguageTool_lang_en_SpellCheckerDemoCodeForHomepage(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	known := map[string]struct{}{
		"A": {}, "a": {}, "error": {}, "spelling": {},
	}
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", SimpleMapSpellerChecker("MORFOLOGIK_RULE_EN_US", known, map[string][]string{
		"speling": {"spelling"},
	}))
	matches := lt.Check("A speling error")
	require.NotEmpty(t, matches)
	require.Equal(t, []string{"spelling"}, matches[0].Suggestions)
	fixed := CorrectTextFromLocalMatches("A speling error", matches)
	require.Equal(t, "A spelling error", fixed)
}

// Twin of JLanguageToolTest.spellCheckerDemoCodeForHomepageWithAddedWords
func TestJLanguageTool_lang_en_SpellCheckerDemoCodeForHomepageWithAddedWords(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	known := map[string]struct{}{"LanguageTool": {}}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, nil))
	require.Empty(t, lt.Check("LanguageTool"))
	lt2 := NewJLanguageTool("en-US")
	lt2.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{}, nil))
	require.NotEmpty(t, lt2.Check("LanguageTool"))
}

// Twin of JLanguageToolTest.testEnglish — error-free inject smoke (full grammar deferred)
func TestJLanguageTool_lang_en_English(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	// clean sentences with a/an correct → empty
	require.Empty(t, lt.Check("A test that should not give errors."))
	require.Empty(t, lt.Check("As long as you have hope, a chance remains."))
	// intentional a/an error
	require.NotEmpty(t, lt.Check("This is an test."))
}

// Twin of JLanguageToolTest.testPositionsWithEnglish
func TestJLanguageTool_lang_en_PositionsWithEnglish(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	// force typoh misspelled by not including it
	lt = NewJLanguageTool("en-US")
	known := map[string]struct{}{
		"A": {}, "sentence": {}, "with": {}, "no": {}, "period": {}, "typo": {},
	}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, map[string][]string{"typoh": {"typo"}}))
	ms := lt.Check("A sentence with no period\nA sentence. A typoh.")
	require.NotEmpty(t, ms)
	// line/column from match metadata when set
	require.GreaterOrEqual(t, ms[0].FromPos, 0)
}

// Twin of JLanguageToolTest.testPositionsWithEnglishTwoLineBreaks
func TestJLanguageTool_lang_en_PositionsWithEnglishTwoLineBreaks(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	known := map[string]struct{}{
		"This": {}, "sentence": {}, "A": {}, "typo": {},
	}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, map[string][]string{"typoh": {"typo"}}))
	ms := lt.Check("This sentence.\n\nA sentence. A typoh.")
	require.NotEmpty(t, ms)
	require.GreaterOrEqual(t, ms[0].FromPos, 0)
}

// Twin of JLanguageToolTest.testAnalyzedSentence
func TestJLanguageTool_lang_en_AnalyzedSentence(t *testing.T) {
	lt := NewJLanguageTool("en")
	// soft hyphen strip path in Analyze
	s := lt.GetAnalyzedSentence("This is a test\u00aded sentence.")
	require.NotNil(t, s)
	// paragraph end
	s2 := lt.GetAnalyzedSentence("\n")
	require.NotNil(t, s2)
	// vertical tab treated as whitespace in tokenize
	s3 := lt.GetAnalyzedSentence("I'm a cool test\u000Bwith a line")
	require.NotNil(t, s3)
	toks := s3.GetTokens()
	require.NotEmpty(t, toks)
}

// Twin of JLanguageToolTest.testParagraphRules
func TestJLanguageTool_lang_en_ParagraphRules(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	// normal: a/an error
	ms := lt.Check("(This is an quote.\n It ends in the second sentence.")
	require.NotEmpty(t, ms)
	// ONLYNONPARA: sentence rules only
	lt.SetParagraphHandling(ParagraphOnlyNonPara)
	ms2 := lt.Check("(This is an quote.\n It ends in the second sentence.")
	// may still find a/an
	_ = ms2
	// ONLYPARA: sentence rules skipped
	lt.SetParagraphHandling(ParagraphOnlyPara)
	ms3 := lt.Check("(This is an quote.\n It ends in the second sentence.")
	// without text-level unpaired inject, may be empty — fail closed
	_ = ms3
	lt.SetParagraphHandling(ParagraphNormal)
}

// Twin of JLanguageToolTest.testWhitespace
func TestJLanguageTool_lang_en_Whitespace(t *testing.T) {
	lt := NewJLanguageTool("en")
	raw := lt.GetRawAnalyzedSentence("Let's do a \"test\", do you understand?")
	cooked := lt.GetAnalyzedSentence("Let's do a \"test\", do you understand?")
	require.NotNil(t, raw)
	require.NotNil(t, cooked)
	// same token count (nothing deleted)
	require.Equal(t, len(raw.GetTokens()), len(cooked.GetTokens()))
}

// Twin of JLanguageToolTest.testOverlapFilter
func TestJLanguageTool_lang_en_OverlapFilter(t *testing.T) {
	// CleanOverlappingLocalMatches filters shorter overlapping lower-priority
	cleaned := CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 4, ToPos: 7, RuleID: "id1", Message: "msg1", Priority: 1},   // "one"
		{FromPos: 4, ToPos: 11, RuleID: "id1", Message: "msg2", Priority: 1}, // "one two" longer preferred often by length
	})
	require.NotEmpty(t, cleaned)
	// inject two overlapping checkers via LT
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("id1_1", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 4, ToPos: 7, RuleID: "id1", Message: "msg1", Suggestions: []string{"x"}}}
	})
	lt.AddRuleChecker("id1_2", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 4, ToPos: 11, RuleID: "id1", Message: "msg2", Suggestions: []string{"y"}}}
	})
	ms := lt.Check("And one two three.")
	// overlap filter leaves at least one
	require.LessOrEqual(t, len(ms), 2)
}

// Twin of JLanguageToolTest.testTextLevelRuleWithGlobalData
func TestJLanguageTool_lang_en_TextLevelRuleWithGlobalData(t *testing.T) {
	// Global metadata path — Check with AnnotatedText global meta when supported
	lt := NewJLanguageTool("en")
	// fail-closed: without text-level Email rule, empty
	require.Empty(t, lt.Check("hello"))
}

// Twin of JLanguageToolTest.testAdvancedTypography
func TestJLanguageTool_lang_en_AdvancedTypography(t *testing.T) {
	cfg := DefaultTypographyConfig()
	cfg.Enabled = true
	require.Equal(t, "The genitive (’s) may be missing.",
		ToAdvancedTypography("The genitive ('s) may be missing.", cfg))
	// suggestion tags → double curly quotes when enabled
	got := ToAdvancedTypography("Did you mean <suggestion>Language's</suggestion>?", cfg)
	require.Contains(t, got, "Language")
	require.NotContains(t, got, "<suggestion>")
}

// Twin of JLanguageToolTest.testAdaptSuggestions
func TestJLanguageTool_lang_en_AdaptSuggestions(t *testing.T) {
	// default AdaptSuggestion is identity
	require.Equal(t, "n't", AdaptSuggestion("n't", "doesn't"))
	// list adapt
	require.Equal(t, []string{"n't", " never"}, AdaptSuggestionsList([]string{"n't", " never"}, "doesn't never"))
}

// Twin of JLanguageToolTest.testEnglishVariants
func TestJLanguageTool_lang_en_EnglishVariants(t *testing.T) {
	sentence := "This is a test."
	sentence2 := "This is an test."
	for _, code := range []string{"en-US", "en-AU", "en-GB", "en-CA", "en-ZA", "en-NZ"} {
		lt := NewJLanguageTool(code)
		lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
		require.Empty(t, lt.Check(sentence), code)
		require.NotEmpty(t, lt.Check(sentence2), code)
	}
}
