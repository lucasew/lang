package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/xx"
	betok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/be"
	"github.com/stretchr/testify/require"
)

// Twin of Belarusian metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id / ignored chars / quotes).
// Java @Deprecated since 3.6 unmaintained; getMaintainedState not overridden.
func TestBelarusianMetadata(t *testing.T) {
	be := NewBelarusian()
	require.Equal(t, "be", be.GetShortCode())
	require.Equal(t, "be", BelarusianShortCode())
	require.Equal(t, "Belarusian", be.GetName())
	require.Equal(t, []string{"BY"}, be.GetCountries())
	// Java len(countries)==1 → "be-BY"
	require.Equal(t, "be-BY", be.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "be/common_words.txt", be.GetCommonWordsPath())
	// Java Belarusian does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, be.GetMaintainedState())
	require.Equal(t, "Alex Buloichik", be.GetMaintainers()[0].Name)
	require.Equal(t, "MORFOLOGIK_RULE_BE_BY", be.GetDefaultSpellingRuleID())
	require.Equal(t, be, Belarusian)

	// getIgnoredCharactersRegex → soft hyphen + combining acute/grave
	re := be.GetIgnoredCharactersRegex()
	require.NotNil(t, re)
	require.True(t, re.MatchString("\u00AD"))
	require.True(t, re.MatchString("\u0301"))
	require.True(t, re.MatchString("\u0300"))

	// Quotes / advanced typography (Java @since 5.1)
	require.Equal(t, "«", be.GetOpeningDoubleQuote())
	require.Equal(t, "»", be.GetClosingDoubleQuote())
	require.Equal(t, "‘", be.GetOpeningSingleQuote())
	require.Equal(t, "’", be.GetClosingSingleQuote())
	require.True(t, be.IsAdvancedTypographyEnabled())
	require.Equal(t, "Скажы «прывітанне»", be.ToAdvancedTypography(`Скажы "прывітанне"`))
}

// Twin of Belarusian.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestBelarusianRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := BelarusianRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE",    // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",              // DoublePunctuationRule
		"MORFOLOGIK_RULE_BE_BY",           // MorfologikBelarusianSpellerRule
		"UPPERCASE_SENTENCE_START",        // UppercaseSentenceStartRule
		"WHITESPACE_RULE",                 // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",             // SentenceWhitespaceRule
		"WHITESPACE_PARAGRAPH",            // WhiteSpaceBeforeParagraphEnd
		"WHITESPACE_PARAGRAPH_BEGIN",      // WhiteSpaceAtBeginOfParagraph
		"TOO_LONG_SENTENCE",               // LongSentenceRule(…, 50)
		"TOO_LONG_PARAGRAPH",              // LongParagraphRule
		"PARAGRAPH_REPEAT_BEGINNING_RULE", // ParagraphRepeatBeginningRule
		"PUNCTUATION_PARAGRAPH_END2",      // PunctuationMarkAtParagraphEnd2
		"BE_SIMPLE_REPLACE",               // SimpleReplaceRule
		"BE_SPECIFIC_CASE",                // BelarusianSpecificCaseRule.getId
	}, ids)
	require.Len(t, ids, 14)
	require.Equal(t, ids, Belarusian.GetRelevantRuleIDs())
	require.Equal(t, ids, NewBelarusian().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	// (Java comments out PunctuationMarkAtParagraphEnd; no unpaired brackets)
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"WHITESPACE_PUNCTUATION", "PUNCTUATION_PARAGRAPH_END",
		"UNPAIRED_BRACKETS", "WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_BE", "BELARUSIAN_SPECIFIC_CASE", // rule getId is BE_SPECIFIC_CASE
		"RUSSIAN_SIMPLE_REPLACE_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Belarusian getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestBelarusian_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Belarusian.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/be/grammar.xml",
	}, files)
	require.Equal(t, files, NewBelarusian().GetRuleFileNamesWithExists(exists))
}

// Twin of Belarusian createDefault* factory wiring → accepted BE twins.
// Java overrides: tagger (DemoTagger), sentence tokenizer, word tokenizer, spelling rule.
// No disambiguator/synthesizer invent (Java has none).
func TestBelarusianCreateDefaultFactories(t *testing.T) {
	be := NewBelarusian()

	// createDefaultTagger → DemoTagger
	tagger := be.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &xx.DemoTagger{}, tagger)

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("be")
	st := be.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Гэта першы сказ. А гэта другі.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → BelarusianWordTokenizer
	wt := be.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &betok.BelarusianWordTokenizer{}, wt)
	toks := wt.Tokenize("камп'ютар")
	require.Equal(t, []string{"камп'ютар"}, toks)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_BE_BY
	require.Equal(t, "MORFOLOGIK_RULE_BE_BY", be.GetDefaultSpellingRuleID())
}

// Twin of Belarusian.getPriorityForId — switch keys as Java writes them.
// Note: Java priority key is BELARUSIAN_SPECIFIC_CASE while rule getId is BE_SPECIFIC_CASE.
func TestBelarusianPriorityForId_OnLang(t *testing.T) {
	be := NewBelarusian()
	require.Equal(t, 10, be.GetPriorityForId("RUSSIAN_SIMPLE_REPLACE_RULE"))
	require.Equal(t, 9, be.GetPriorityForId("BELARUSIAN_SPECIFIC_CASE"))
	require.Equal(t, -1, be.GetPriorityForId("Word_root_repeat"))
	require.Equal(t, -2, be.GetPriorityForId("PUNCT_DPT_2"))
	require.Equal(t, -15, be.GetPriorityForId("TOO_LONG_PARAGRAPH"))
	// BE_SPECIFIC_CASE (actual rule id) is not in the Java switch → base Language
	require.Equal(t, 0, be.GetPriorityForId("BE_SPECIFIC_CASE"))
	require.Equal(t, 0, be.GetPriorityForId("UNKNOWN_BE_XYZ"))
}
