package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	crhsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/crh"
	crhtag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/crh"
	crhtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/crh"
	"github.com/stretchr/testify/require"
)

// Twin of CrimeanTatar metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / speller getId).
// Java overrides getMaintainedState → ActivelyMaintained.
func TestCrimeanTatarMetadata(t *testing.T) {
	crh := NewCrimeanTatar()
	require.Equal(t, "crh", crh.GetShortCode())
	require.Equal(t, "crh", CrimeanTatarShortCode())
	require.Equal(t, "Crimean Tatar", crh.GetName())
	require.Equal(t, []string{"UA"}, crh.GetCountries())
	// Java len(countries)==1 → "crh-UA"
	require.Equal(t, "crh-UA", crh.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "crh/common_words.txt", crh.GetCommonWordsPath())
	// Java CrimeanTatar overrides getMaintainedState → ActivelyMaintained
	require.Equal(t, languagetool.ActivelyMaintained, crh.GetMaintainedState())
	require.Equal(t, "Andriy Rysin", crh.GetMaintainers()[0].Name)
	require.Equal(t, "MORFOLOGIK_RULE_CRH_UA", crh.GetDefaultSpellingRuleID())
	require.Equal(t, crh, CrimeanTatar)
}

// Twin of CrimeanTatar.getRelevantRules — class getId order exactly as Java Arrays.asList.
// No SharedLayout invent (no unpaired brackets, empty line, long paragraph, word-repeat, …).
func TestCrimeanTatarRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := CrimeanTatarRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"WHITESPACE_PARAGRAPH",         // WhiteSpaceBeforeParagraphEnd
		"WHITESPACE_PARAGRAPH_BEGIN",   // WhiteSpaceAtBeginOfParagraph
		"MORFOLOGIK_RULE_CRH_UA",       // MorfologikCrimeanTatarSpellerRule
	}, ids)
	require.Len(t, ids, 8)
	require.Equal(t, ids, CrimeanTatar.GetRelevantRuleIDs())
	require.Equal(t, ids, NewCrimeanTatar().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "TOO_LONG_SENTENCE",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"UNPAIRED_BRACKETS", "WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_CRH", "MORFOLOGIK_RULE_CRH_TR",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of CrimeanTatar getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestCrimeanTatar_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := CrimeanTatar.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/crh/grammar.xml",
	}, files)
	require.Equal(t, files, NewCrimeanTatar().GetRuleFileNamesWithExists(exists))
}

// Twin of CrimeanTatar factory wiring → accepted CRH twins.
// Java: getSentenceTokenizer, getWordTokenizer, getTagger (lazy), createDefaultSynthesizer.
// No createDefaultDisambiguator / createDefaultSpellingRule invent.
func TestCrimeanTatarCreateDefaultFactories(t *testing.T) {
	crh := NewCrimeanTatar()

	// getTagger → CrimeanTatarTagger (Java path /crh/crimean_tatar.dict)
	tagger := crh.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &crhtag.CrimeanTatarTagger{}, tagger)
	require.Equal(t, crhtag.CrimeanTatarTaggerDictPath, tagger.GetDictionaryPath())
	require.NotNil(t, tagger.GetWordTagger())

	// getSentenceTokenizer → SRXSentenceTokenizer("crh")
	st := crh.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Bu birinci cümle. Ve bu ikinci.")
	require.GreaterOrEqual(t, len(sents), 2)

	// getWordTokenizer → CrimeanTatarWordTokenizer
	wt := crh.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &crhtok.CrimeanTatarWordTokenizer{}, wt)
	toks := wt.Tokenize("qırım-tatar")
	require.NotEmpty(t, toks)

	// createDefaultSynthesizer → CrimeanTatarSynthesizer.INSTANCE
	synth := crh.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &crhsynth.CrimeanTatarSynthesizer{}, synth)
	require.Same(t, crhsynth.Instance(), synth)
	require.Same(t, crhsynth.INSTANCE, synth)
	require.Equal(t, crhsynth.CrimeanTatarSynthDict, synth.ResourceFileName)

	// speller getId → MORFOLOGIK_RULE_CRH_UA
	require.Equal(t, "MORFOLOGIK_RULE_CRH_UA", crh.GetDefaultSpellingRuleID())
}
