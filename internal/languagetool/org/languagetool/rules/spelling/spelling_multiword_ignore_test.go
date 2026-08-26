package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAddIgnoreWords_MultiTokenGoesToMultiWordIgnore(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddIgnoreWords("Microsoft Entra", "log4j", "fit2work")
	require.Contains(t, r.IgnoreWords, "log4j")
	require.Contains(t, r.IgnoreWords, "fit2work")
	require.NotContains(t, r.IgnoreWords, "Microsoft Entra")
	require.Len(t, r.MultiWordIgnore, 1)
	require.Equal(t, []string{"Microsoft", "Entra"}, r.MultiWordIgnore[0])
}

func TestDefaultTokenizeIgnoreLine_WordTokenizerPunctuation(t *testing.T) {
	// Java WordTokenizer splits on '.'; EnglishWordTokenizer may keep "P." as one token.
	got := DefaultTokenizeIgnoreLine("en", "P. Sherman 42 Wallaby Way")
	require.NotEmpty(t, got)
	// Must produce multiple tokens (not a single Fields-like blob).
	require.Greater(t, len(got), 1)
	// Last tokens include Wallaby Way
	require.Contains(t, got, "Wallaby")
	require.Contains(t, got, "Way")
}

func TestMarkMultiWordIgnoreSpelling(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddIgnoreWords("Microsoft Entra")
	// Match Analyze path tokenizer (JLanguageTool.Analyze uses WordTokenizerForLanguage).
	sent := languagetool.AnalyzeWithTokenizer(
		"Use Microsoft Entra today.",
		languagetool.WordTokenizerForLanguage("en"),
	)
	r.MarkMultiWordIgnoreSpelling(sent)
	var marked []string
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.IsIgnoredBySpeller() {
			marked = append(marked, tok.GetToken())
		}
	}
	require.Equal(t, []string{"Microsoft", "Entra"}, marked)
}

func TestMarkMultiWordIgnoreSpelling_NoPartialMatch(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddIgnoreWords("Microsoft Entra")
	sent := languagetool.AnalyzeWithTokenizer(
		"Microsoft alone.",
		languagetool.WordTokenizerForLanguage("en"),
	)
	r.MarkMultiWordIgnoreSpelling(sent)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "Microsoft" {
			require.False(t, tok.IsIgnoredBySpeller())
		}
	}
}

func TestApplyDefaultSpellingWordLists_GlobalMultiWord(t *testing.T) {
	if DiscoverSpellingGlobal() == "" {
		t.Skip("spelling_global.txt not in tree")
	}
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	ApplyDefaultSpellingWordLists(r)
	// "Microsoft Entra" is in official spelling_global.txt
	found := false
	for _, p := range r.MultiWordIgnore {
		if len(p) >= 2 && p[0] == "Microsoft" && p[1] == "Entra" {
			found = true
		}
	}
	require.True(t, found, "expected Microsoft Entra multi-word ignore from spelling_global.txt, got %+v", r.MultiWordIgnore)
}

func TestDefaultTokenizeIgnoreLine_MatchesAnalyzeEnglish(t *testing.T) {
	phrase := "Microsoft Entra"
	toks := DefaultTokenizeIgnoreLine("en", phrase)
	sent := languagetool.AnalyzeWithTokenizer(phrase+".", languagetool.WordTokenizerForLanguage("en"))
	var surfaces []string
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		// drop sentence-end punctuation token if separate
		if tok.GetToken() == "." {
			continue
		}
		surfaces = append(surfaces, tok.GetToken())
	}
	require.Equal(t, toks, surfaces)
}

func TestAddIgnoreWords_DisableTokenizeNewWords(t *testing.T) {
	// Java tokenizeNewWords()=false: multi-token line → single IgnoreWords key, no antipattern
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_CA_ES", "spell", "ca")
	r.DisableTokenizeNewWords = true
	r.AddIgnoreWords("Microsoft Entra", "log4j")
	require.Contains(t, r.IgnoreWords, "Microsoft Entra")
	require.Contains(t, r.IgnoreWords, "log4j")
	require.Empty(t, r.MultiWordIgnore)
	require.Empty(t, r.AntiPatterns)
}

func TestAddIgnoreWords_TokenizeNewWordsDefault(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_NL_NL", "spell", "nl")
	r.AddIgnoreWords("Microsoft Entra")
	require.NotContains(t, r.IgnoreWords, "Microsoft Entra")
	require.Len(t, r.MultiWordIgnore, 1)
}
