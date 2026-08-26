package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	sksynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/sk"
	sktag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/sk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of Slovak metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
func TestSlovakMetadata(t *testing.T) {
	sk := NewSlovak()
	require.Equal(t, "sk", sk.GetShortCode())
	require.Equal(t, "sk", SlovakShortCode())
	require.Equal(t, "Slovak", sk.GetName())
	require.Equal(t, []string{"SK"}, sk.GetCountries())
	// Java len(countries)==1 → "sk-SK"
	require.Equal(t, "sk-SK", sk.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "sk/common_words.txt", sk.GetCommonWordsPath())
	// Java Slovak does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, sk.GetMaintainedState())
	require.Equal(t, "Zdenko Podobný", sk.GetMaintainers()[0].Name)
	require.Equal(t, "http://sk-spell.sk.cx", sk.GetMaintainers()[0].URL)
	require.Equal(t, "MORFOLOGIK_RULE_SK_SK", sk.GetDefaultSpellingRuleID())
	require.Equal(t, sk, Slovak)
}

// Twin of Slovak.getRelevantRules — class getId order exactly as Java Arrays.asList
// (SlovakVesRule commented out — omitted).
func TestSlovakRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := SlovakRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SK_COMPOUNDS",                 // CompoundRule
		"MORFOLOGIK_RULE_SK_SK",        // MorfologikSlovakSpellerRule
	}, ids)
	require.Equal(t, ids, Slovak.GetRelevantRuleIDs())
	require.Equal(t, ids, NewSlovak().GetRelevantRuleIDs())
	// invent / SharedLayout extras / commented SlovakVesRule must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"MORFOLOGIK_RULE_SK", "SK_UNPAIRED_BRACKETS", "SK_VES_RULE", "SK_VES",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Slovak.getRuleFileNames — base grammar.xml + grammar-typography.xml.
func TestSlovak_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Slovak.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/sk/grammar.xml",
		"/org/languagetool/rules/sk/grammar-typography.xml",
	}, files)
	require.Equal(t, files, NewSlovak().GetRuleFileNamesWithExists(exists))
}

// Twin of Slovak createDefault* factory wiring → real accepted SK twins.
func TestSlovakCreateDefaultFactories(t *testing.T) {
	sk := NewSlovak()

	// createDefaultTagger → SlovakTagger over official slovak.dict
	require.NotEmpty(t, sktag.DiscoverSlovakPOSDict(),
		"official slovak.dict must be discoverable")
	tagger := sk.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &sktag.SlovakTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	require.Same(t, sktag.DefaultSlovakTagger, tagger)
	// Spot-check a known tagged form (same surface SlovakTaggerTest uses)
	readings := tagger.Tag([]string{"slova"})
	require.NotEmpty(t, readings)
	require.True(t, readings[0].IsTagged(), "SlovakTagger must tag 'slova' from slovak.dict")

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("sk")
	st := sk.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Toto je prvá veta. A toto je druhá.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → Language default WordTokenizer (no SK override)
	wt := sk.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &tokenizers.WordTokenizer{}, wt)
	toks := wt.Tokenize("Toto je test.")
	require.Contains(t, toks, "Toto")
	require.Contains(t, toks, "test")

	// createDefaultSynthesizer → SlovakSynthesizer.INSTANCE
	require.NotEmpty(t, sksynth.DiscoverSlovakSynthDict(),
		"official slovak_synth.dict must be discoverable")
	synth := sk.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &sksynth.SlovakSynthesizer{}, synth)
	require.Same(t, sksynth.Instance(), synth)
	require.Same(t, sksynth.INSTANCE, synth)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_SK_SK
	require.Equal(t, "MORFOLOGIK_RULE_SK_SK", sk.GetDefaultSpellingRuleID())
}
