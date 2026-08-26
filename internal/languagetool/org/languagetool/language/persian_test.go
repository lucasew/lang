package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of Persian metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState; no createDefaultSpellingRule).
func TestPersianMetadata(t *testing.T) {
	fa := NewPersian()
	require.Equal(t, "fa", fa.GetShortCode())
	require.Equal(t, "fa", PersianShortCode())
	require.Equal(t, "Persian", fa.GetName())
	require.Equal(t, []string{"IR", "AF"}, fa.GetCountries())
	// Java len(countries)==2 → bare "fa" (not fa-IR; first country is locale only)
	require.Equal(t, "fa", fa.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "fa/common_words.txt", fa.GetCommonWordsPath())
	// Java Persian does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, fa.GetMaintainedState())
	ms := fa.GetMaintainers()
	require.Len(t, ms, 3)
	require.Equal(t, "Reza1615", ms[0].Name)
	require.Equal(t, "Alireza Eskandarpour Shoferi", ms[1].Name)
	require.Equal(t, "Ebrahim Byagowi", ms[2].Name)
	// No invent: no createDefaultSpellingRule / no speller in getRelevantRules
	require.Empty(t, fa.GetDefaultSpellingRuleID())
	require.Equal(t, fa, Persian)
}

// Twin of Persian.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestPersianRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := PersianRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE",         // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",                   // DoublePunctuationRule
		"WHITESPACE_RULE",                      // MultipleWhitespaceRule
		"TOO_LONG_SENTENCE",                    // LongSentenceRule(…, 50)
		"PERSIAN_COMMA_PARENTHESIS_WHITESPACE", // PersianCommaWhitespaceRule
		"PERSIAN_DOUBLE_PUNCTUATION",           // PersianDoublePunctuationRule
		"PERSIAN_WORD_REPEAT_BEGINNING_RULE",   // PersianWordRepeatBeginningRule
		"PERSIAN_WORD_REPEAT_RULE",             // PersianWordRepeatRule
		"FA_SIMPLE_REPLACE",                    // SimpleReplaceRule
		"FA_SPACE_BEFORE_CONJUNCTION",          // PersianSpaceBeforeRule
		"FA_WORD_COHERENCY",                    // WordCoherencyRule
	}, ids)
	require.Len(t, ids, 11)
	require.Equal(t, ids, Persian.GetRelevantRuleIDs())
	require.Equal(t, ids, NewPersian().GetRelevantRuleIDs())
	// invent / SharedLayout extras / speller IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"UNPAIRED_BRACKETS", "UPPERCASE_SENTENCE_START",
		"TOO_LONG_SENTENCE_FA", "HUNSPELL_RULE", "MORFOLOGIK_RULE_FA",
		"MORFOLOGIK_RULE_FA_IR", "HUNSPELL_RULE_FA",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Persian getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestPersian_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Persian.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/fa/grammar.xml",
	}, files)
	require.Equal(t, files, NewPersian().GetRuleFileNamesWithExists(exists))
}

// Twin of Persian createDefault* factory wiring.
// Java overrides: sentence tokenizer, word tokenizer only.
// No tagger/speller/disambiguator/synthesizer invent.
func TestPersianCreateDefaultFactories(t *testing.T) {
	fa := NewPersian()

	// createDefaultTagger → Language default DemoTagger (no Persian override)
	tagger := fa.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &xx.DemoTagger{}, tagger)

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("fa")
	st := fa.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("این جمله اول است. این جمله دوم است.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → PersianWordTokenizer
	wt := fa.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &tokenizers.PersianWordTokenizer{}, wt)
	toks := wt.Tokenize("سلام، دنیا؟")
	require.Contains(t, toks, "سلام")
	require.Contains(t, toks, "،")
	require.Contains(t, toks, "دنیا")
	require.Contains(t, toks, "؟")

	// no createDefaultSpellingRule
	require.Empty(t, fa.GetDefaultSpellingRuleID())
}
