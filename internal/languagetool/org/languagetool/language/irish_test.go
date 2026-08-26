package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	gasynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ga"
	gadis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ga"
	gatag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Twin of Irish metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id / getDefaultLanguageVariant).
// Java overrides getMaintainedState → ActivelyMaintained; countries {"IE"}.
func TestIrishMetadata(t *testing.T) {
	ga := NewIrish()
	require.Equal(t, "ga", ga.GetShortCode())
	require.Equal(t, "ga", IrishShortCode())
	require.Equal(t, "Irish", ga.GetName())
	require.Equal(t, []string{"IE"}, ga.GetCountries())
	// Java len(countries)==1 → "ga-IE"
	require.Equal(t, "ga-IE", ga.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "ga/common_words.txt", ga.GetCommonWordsPath())
	// Java Irish overrides getMaintainedState → ActivelyMaintained
	require.Equal(t, languagetool.ActivelyMaintained, ga.GetMaintainedState())
	ms := ga.GetMaintainers()
	require.Len(t, ms, 4)
	require.Equal(t, "Jim O'Regan", ms[0].Name)
	require.Equal(t, "Emily Barnes", ms[1].Name)
	require.Equal(t, "Mícheál J. Ó Meachair", ms[2].Name)
	require.Equal(t, "Seanán Ó Coistín", ms[3].Name)
	require.Equal(t, "MORFOLOGIK_RULE_GA_IE", ga.GetDefaultSpellingRuleID())
	// getDefaultLanguageVariant → getInstance() self (registry short code "ga")
	require.Equal(t, "ga", ga.GetDefaultLanguageVariantCode())
	require.Equal(t, "ga", IrishDefaultLanguageVariantCode())
	require.Equal(t, ga, Irish)
}

// Twin of Irish.getRelevantRules — class getId order exactly as Java Arrays.asList.
// Note: UppercaseSentenceStartRule appears twice in live Java list (duplicate registration).
// Pattern grammar / IrishPatternRuleTest: honest gap (pattern sector).
func TestIrishRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := IrishRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (custom pairs)
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule (1st)
		"TOO_LONG_SENTENCE",            // LongSentenceRule (defaultLength 50)
		"TOO_LONG_PARAGRAPH",           // LongParagraphRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule (2nd — Java duplicate)
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"WHITESPACE_PARAGRAPH",         // WhiteSpaceBeforeParagraphEnd
		"WHITESPACE_PARAGRAPH_BEGIN",   // WhiteSpaceAtBeginOfParagraph
		"PARAGRAPH_REPEAT_BEGINNING_RULE", // ParagraphRepeatBeginningRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"MORFOLOGIK_RULE_GA_IE",        // MorfologikIrishSpellerRule
		"GA_LOGAINM",                   // LogainmRule
		"GA_PEOPLE",                    // PeopleRule
		"GA_SPASANNA",                  // SpacesRule
		"GA_COMPOUNDS",                 // CompoundRule
		"GA_PRESTANDARD_REPLACE",       // PrestandardReplaceRule
		"GA_REPLACE",                   // IrishReplaceRule
		"GA_FGB_EQ_REPLACE",            // IrishFGBEqReplaceRule
		"GA_ENGLISH_HOMOPHONE",         // EnglishHomophoneRule
		"GA_DHA_NO_BEIRT",              // DhaNoBeirtRule
		"GA_DATIVE_PLURALS_STD",        // DativePluralStandardReplaceRule
		"GA_SPECIFIC_CASE",             // IrishSpecificCaseRule
	}, ids)
	require.Len(t, ids, 25)
	nUpper := 0
	for _, id := range ids {
		if id == "UPPERCASE_SENTENCE_START" {
			nUpper++
		}
	}
	require.Equal(t, 2, nUpper, "Java asList registers UppercaseSentenceStartRule twice")
	require.Equal(t, ids, Irish.GetRelevantRuleIDs())
	require.Equal(t, ids, NewIrish().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"WHITESPACE_PUNCTUATION", "HUNSPELL_RULE", "MORFOLOGIK_RULE_GA",
		"GA_UNPAIRED_BRACKETS",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Irish.getPriorityForId — TOO_LONG_PARAGRAPH → -15 else Language super.
func TestIrish_GetPriorityForId(t *testing.T) {
	ga := NewIrish()
	require.Equal(t, -15, ga.GetPriorityForId("TOO_LONG_PARAGRAPH"))
	require.Equal(t, -15, IrishPriorityForId("TOO_LONG_PARAGRAPH"))
	// base Language still applies for STYLE
	require.Equal(t, -50, ga.GetPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, ga.GetPriorityForId("UNKNOWN_GA_XYZ"))
}

// Twin of Irish getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestIrish_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Irish.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/ga/grammar.xml",
	}, files)
	require.Equal(t, files, NewIrish().GetRuleFileNamesWithExists(exists))
}

// Twin of Irish createDefault* factory wiring → accepted GA twins.
// Java overrides: tagger, sentence tokenizer, word tokenizer, disambiguator,
// synthesizer, spelling rule. Pattern grammar load: honest gap (pattern sector).
func TestIrishCreateDefaultFactories(t *testing.T) {
	ga := NewIrish()

	// createDefaultTagger → IrishTagger over official irish.dict
	require.NotEmpty(t, gatag.DiscoverIrishPOSDict(),
		"official irish.dict must be discoverable")
	tagger := ga.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &gatag.IrishTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	require.Same(t, gatag.DefaultIrishTagger, tagger)
	// Spot-check a known tagged form (IrishTaggerTest surface: anseo → Adv:Loc)
	readings := tagger.Tag([]string{"anseo"})
	require.NotEmpty(t, readings)
	require.True(t, readings[0].IsTagged(), "IrishTagger must tag 'anseo' from irish.dict")

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("ga")
	st := ga.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Is é seo an chéad abairt. Agus seo an dara.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → core WordTokenizer (Java new WordTokenizer())
	wt := ga.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &tokenizers.WordTokenizer{}, wt)
	toks := wt.Tokenize("Is é seo tástáil.")
	require.Contains(t, toks, "Is")
	require.Contains(t, toks, "tástáil")

	// createDefaultDisambiguator → IrishHybridDisambiguator
	dis := ga.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &gadis.IrishHybridDisambiguator{}, dis)
	// Stages wired when official multiwords / disambiguation.xml discoverable
	if gadis.DiscoverIrishMultiwords() != "" {
		require.NotNil(t, dis.Chunker, "Java MultiWordChunker.getInstance(/ga/multiwords.txt)")
	}
	if gadis.DiscoverIrishDisambiguationXML() != "" {
		require.NotNil(t, dis.Rules, "Java XmlRuleDisambiguator(Irish.getInstance())")
	}

	// createDefaultSynthesizer → IrishSynthesizer.INSTANCE
	require.NotEmpty(t, gasynth.DiscoverIrishSynthDict(),
		"official irish_synth.dict must be discoverable")
	synth := ga.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &gasynth.IrishSynthesizer{}, synth)
	require.Same(t, gasynth.Instance(), synth)
	require.Same(t, gasynth.INSTANCE, synth)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_GA_IE
	require.Equal(t, "MORFOLOGIK_RULE_GA_IE", ga.GetDefaultSpellingRuleID())
}
