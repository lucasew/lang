package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of Slovenian metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
// Java @Deprecated since 3.6 unmaintained; getMaintainedState not overridden.
func TestSlovenianMetadata(t *testing.T) {
	sl := NewSlovenian()
	require.Equal(t, "sl", sl.GetShortCode())
	require.Equal(t, "sl", SlovenianShortCode())
	require.Equal(t, "Slovenian", sl.GetName())
	require.Equal(t, []string{"SI"}, sl.GetCountries())
	// Java len(countries)==1 → "sl-SI"
	require.Equal(t, "sl-SI", sl.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "sl/common_words.txt", sl.GetCommonWordsPath())
	// Java Slovenian does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, sl.GetMaintainedState())
	require.Equal(t, "Martin Srebotnjak", sl.GetMaintainers()[0].Name)
	require.Equal(t, "MORFOLOGIK_RULE_SL_SI", sl.GetDefaultSpellingRuleID())
	require.Equal(t, sl, Slovenian)
}

// Twin of Slovenian.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestSlovenianRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := SlovenianRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule
		"MORFOLOGIK_RULE_SL_SI",        // MorfologikSlovenianSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
	}, ids)
	require.Equal(t, ids, Slovenian.GetRelevantRuleIDs())
	require.Equal(t, ids, NewSlovenian().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"MORFOLOGIK_RULE_SL", "SL_UNPAIRED_BRACKETS",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Slovenian getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestSlovenian_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Slovenian.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/sl/grammar.xml",
	}, files)
	require.Equal(t, files, NewSlovenian().GetRuleFileNamesWithExists(exists))
}

// Twin of Slovenian createDefault* factory wiring (Java overrides only sentence tokenizer
// + spelling rule; no tagger/synthesizer invent).
func TestSlovenianCreateDefaultFactories(t *testing.T) {
	sl := NewSlovenian()

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("sl")
	st := sl.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("To je prvi stavek. In to je drugi.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → Language default WordTokenizer (no SL override)
	wt := sl.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &tokenizers.WordTokenizer{}, wt)
	toks := wt.Tokenize("To je test.")
	require.Contains(t, toks, "To")
	require.Contains(t, toks, "test")

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_SL_SI
	require.Equal(t, "MORFOLOGIK_RULE_SL_SI", sl.GetDefaultSpellingRuleID())
}
