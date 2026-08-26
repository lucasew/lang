package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tltag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/tl"
	tltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/tl"
	"github.com/stretchr/testify/require"
)

// Twin of Tagalog metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
// Java @Deprecated since 3.6 unmaintained; getMaintainedState not overridden.
func TestTagalogMetadata(t *testing.T) {
	tl := NewTagalog()
	require.Equal(t, "tl", tl.GetShortCode())
	require.Equal(t, "tl", TagalogShortCode())
	require.Equal(t, "Tagalog", tl.GetName())
	require.Equal(t, []string{"PH"}, tl.GetCountries())
	// Java len(countries)==1 → "tl-PH"
	require.Equal(t, "tl-PH", tl.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "tl/common_words.txt", tl.GetCommonWordsPath())
	// Java Tagalog does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, tl.GetMaintainedState())
	require.Equal(t, "Nathaniel Oco", tl.GetMaintainers()[0].Name)
	require.Equal(t, "Allan Borra", tl.GetMaintainers()[1].Name)
	require.Equal(t, "MORFOLOGIK_RULE_TL", tl.GetDefaultSpellingRuleID())
	require.Equal(t, tl, Tagalog)
}

// Twin of Tagalog.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestTagalogRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := TagalogRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"MORFOLOGIK_RULE_TL",           // MorfologikTagalogSpellerRule
	}, ids)
	require.Equal(t, ids, Tagalog.GetRelevantRuleIDs())
	require.Equal(t, ids, NewTagalog().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"MORFOLOGIK_RULE_TL_PH", "TL_UNPAIRED_BRACKETS", "WORD_REPEAT_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Tagalog getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestTagalog_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Tagalog.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/tl/grammar.xml",
	}, files)
	require.Equal(t, files, NewTagalog().GetRuleFileNamesWithExists(exists))
}

// Twin of Tagalog createDefault* factory wiring → real accepted TL twins.
// Java overrides: sentence tokenizer, word tokenizer, tagger, spelling rule.
// No synthesizer/disambiguator invent (Java has none).
func TestTagalogCreateDefaultFactories(t *testing.T) {
	tl := NewTagalog()

	// createDefaultTagger → TagalogTagger over official tagalog.dict
	require.NotEmpty(t, tltag.DiscoverTagalogPOSDict(),
		"official tagalog.dict must be discoverable")
	tagger := tl.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &tltag.TagalogTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	require.Same(t, tltag.DefaultTagalogTagger, tagger)
	// Spot-check a known tagged form (same surface TagalogTaggerTest uses)
	readings := tagger.Tag([]string{"ako"})
	require.NotEmpty(t, readings)
	require.True(t, readings[0].IsTagged(), "TagalogTagger must tag 'ako' from tagalog.dict")

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("tl")
	st := tl.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Ito ay unang pangungusap. At ito ang pangalawa.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → TagalogWordTokenizer (hyphen is a token delimiter)
	wt := tl.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &tltok.TagalogWordTokenizer{}, wt)
	toks := wt.Tokenize("araw-araw")
	require.Contains(t, toks, "araw")
	require.Contains(t, toks, "-")

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_TL
	require.Equal(t, "MORFOLOGIK_RULE_TL", tl.GetDefaultSpellingRuleID())
}
