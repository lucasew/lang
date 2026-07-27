package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	asttag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ast"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of Asturian metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
// Java @Deprecated since 3.6 unmaintained; getMaintainedState not overridden.
func TestAsturianMetadata(t *testing.T) {
	ast := NewAsturian()
	require.Equal(t, "ast", ast.GetShortCode())
	require.Equal(t, "ast", AsturianShortCode())
	require.Equal(t, "Asturian", ast.GetName())
	require.Equal(t, []string{"ES"}, ast.GetCountries())
	// Java len(countries)==1 → "ast-ES"
	require.Equal(t, "ast-ES", ast.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "ast/common_words.txt", ast.GetCommonWordsPath())
	// Java Asturian does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, ast.GetMaintainedState())
	require.Equal(t, "Xesús González Rato", ast.GetMaintainers()[0].Name)
	require.Equal(t, "MORFOLOGIK_RULE_AST", ast.GetDefaultSpellingRuleID())
	require.Equal(t, ast, Asturian)
}

// Twin of Asturian.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestAsturianRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := AsturianRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule
		"MORFOLOGIK_RULE_AST",          // MorfologikAsturianSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
	}, ids)
	require.Equal(t, ids, Asturian.GetRelevantRuleIDs())
	require.Equal(t, ids, NewAsturian().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"MORFOLOGIK_RULE_AST_ES", "AST_UNPAIRED_BRACKETS", "WORD_REPEAT_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Asturian getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestAsturian_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Asturian.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/ast/grammar.xml",
	}, files)
	require.Equal(t, files, NewAsturian().GetRuleFileNamesWithExists(exists))
}

// Twin of Asturian createDefault* factory wiring → real accepted AST twins.
// Java overrides: sentence tokenizer, tagger, spelling rule.
// No word tokenizer / synthesizer / disambiguator invent (Java has none).
func TestAsturianCreateDefaultFactories(t *testing.T) {
	ast := NewAsturian()

	// createDefaultTagger → AsturianTagger (Java path /ast/asturian.dict)
	tagger := ast.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &asttag.AsturianTagger{}, tagger)
	require.Equal(t, asttag.AsturianTaggerDictPath, tagger.GetDictionaryPath())
	require.NotNil(t, tagger.GetWordTagger())

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("ast")
	st := ast.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Apega testu equí. Y equí hai otru.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → Language default WordTokenizer (no AST override)
	wt := ast.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &tokenizers.WordTokenizer{}, wt)
	toks := wt.Tokenize("Esto ye un test.")
	require.Contains(t, toks, "Esto")
	require.Contains(t, toks, "test")

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_AST
	require.Equal(t, "MORFOLOGIK_RULE_AST", ast.GetDefaultSpellingRuleID())
}
