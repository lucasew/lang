package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	datag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/da"
	dadis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/da"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of Danish metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
// Java @Deprecated since 3.6 unmaintained; getMaintainedState not overridden.
func TestDanishMetadata(t *testing.T) {
	da := NewDanish()
	require.Equal(t, "da", da.GetShortCode())
	require.Equal(t, "da", DanishShortCode())
	require.Equal(t, "Danish", da.GetName())
	require.Equal(t, []string{"DK"}, da.GetCountries())
	// Java len(countries)==1 → "da-DK"
	require.Equal(t, "da-DK", da.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "da/common_words.txt", da.GetCommonWordsPath())
	// Java Danish does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, da.GetMaintainedState())
	require.Equal(t, "Esben Aaberg", da.GetMaintainers()[0].Name)
	require.Equal(t, "Henrik Bendt", da.GetMaintainers()[1].Name)
	require.Equal(t, "HUNSPELL_RULE", da.GetDefaultSpellingRuleID())
	require.Equal(t, da, Danish)
}

// Twin of Danish.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestDanishRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := DanishRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (custom start/end)
		"HUNSPELL_RULE",                // HunspellRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		// WORD_REPEAT_RULE implemented in grammar.xml (Java comment) — not class-based
		"WHITESPACE_RULE", // MultipleWhitespaceRule
	}, ids)
	require.Len(t, ids, 6)
	require.Equal(t, ids, Danish.GetRelevantRuleIDs())
	require.Equal(t, ids, NewDanish().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"WORD_REPEAT_RULE", "DA_UNPAIRED_BRACKETS", "MORFOLOGIK_RULE_DA",
		"MORFOLOGIK_RULE_DA_DK",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Danish getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestDanish_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Danish.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/da/grammar.xml",
	}, files)
	require.Equal(t, files, NewDanish().GetRuleFileNamesWithExists(exists))
}

// Twin of Danish createDefault* factory wiring → accepted DA twins.
// Java overrides: tagger, sentence tokenizer, disambiguator, spelling rule.
// No word-tokenizer / synthesizer invent (Java has none).
func TestDanishCreateDefaultFactories(t *testing.T) {
	da := NewDanish()

	// createDefaultTagger → DanishTagger over official danish.dict
	require.NotEmpty(t, datag.DiscoverDanishPOSDict(),
		"official danish.dict must be discoverable")
	tagger := da.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &datag.DanishTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	require.Same(t, datag.DefaultDanishTagger, tagger)

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("da")
	st := da.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Dette er den første sætning. Og dette er den anden.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → Language default WordTokenizer (no DA override)
	wt := da.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &tokenizers.WordTokenizer{}, wt)
	toks := wt.Tokenize("Dette er en test.")
	require.Contains(t, toks, "Dette")
	require.Contains(t, toks, "test")

	// createDefaultDisambiguator → XmlRuleDisambiguator(Danish)
	require.NotEmpty(t, dadis.DiscoverDanishDisambiguationXML(),
		"official da/disambiguation.xml must be discoverable")
	dis := da.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &disambigrules.XmlRuleDisambiguator{}, dis)
	require.NotEmpty(t, dis.Rules, "XmlRuleDisambiguator(Danish) must load rules")
	require.Same(t, dadis.DanishXmlRuleDisambiguator(), dis)

	// createDefaultSpellingRule id → HUNSPELL_RULE
	require.Equal(t, "HUNSPELL_RULE", da.GetDefaultSpellingRuleID())
}
