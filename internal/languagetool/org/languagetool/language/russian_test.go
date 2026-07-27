package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	rusynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ru"
	rudis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ru"
	rutag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	rutok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ru"
	"github.com/stretchr/testify/require"
)

// Twin of Russian metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / quotes / typography / createDefaultSpellingRule id /
// getIgnoredCharactersRegex).
func TestRussianMetadata(t *testing.T) {
	ru := NewRussian()
	require.Equal(t, "ru", ru.GetShortCode())
	require.Equal(t, "ru", RussianShortCode())
	require.Equal(t, "Russian", ru.GetName())
	require.Equal(t, []string{"RU"}, ru.GetCountries())
	// Java len(countries)==1 → "ru-RU"
	require.Equal(t, "ru-RU", ru.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "ru/common_words.txt", ru.GetCommonWordsPath())
	require.Equal(t, languagetool.ActivelyMaintained, ru.GetMaintainedState())
	require.Equal(t, "Yakov Reztsov", ru.GetMaintainers()[0].Name)
	require.Equal(t, "http://myooo.ru/content/view/83/43/", ru.GetMaintainers()[0].URL)
	require.Equal(t, "MORFOLOGIK_RULE_RU_RU", ru.GetDefaultSpellingRuleID())
	require.Equal(t, ru, Russian)

	// Quotes + advanced typography (Java @since 5.1)
	require.Equal(t, "«", ru.GetOpeningDoubleQuote())
	require.Equal(t, "»", ru.GetClosingDoubleQuote())
	require.Equal(t, "‘", ru.GetOpeningSingleQuote())
	require.Equal(t, "’", ru.GetClosingSingleQuote())
	require.True(t, ru.IsAdvancedTypographyEnabled())

	// getIgnoredCharactersRegex → soft hyphen + acute/grave combining marks
	re := ru.GetIgnoredCharactersRegex()
	require.NotNil(t, re)
	require.True(t, re.MatchString("\u00AD"))
	require.True(t, re.MatchString("\u0301"))
	require.True(t, re.MatchString("\u0300"))
}

// Twin of Russian.getRelevantRules — class getId order exactly as Java Arrays.asList
// (commented-out DoublePunctuation / WordRepeat / EmptyLine / PunctuationMarkAtParagraphEnd /
// Readability pair omitted).
func TestRussianRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := RussianRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		// DoublePunctuationRule commented out
		"UPPERCASE_SENTENCE_START", // UppercaseSentenceStartRule
		"MORFOLOGIK_RULE_RU_RU",    // MorfologikRussianSpellerRule
		// WordRepeatRule commented out
		"WHITESPACE_RULE",                 // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",             // SentenceWhitespaceRule
		"WHITESPACE_PARAGRAPH",            // WhiteSpaceBeforeParagraphEnd
		"WHITESPACE_PARAGRAPH_BEGIN",      // WhiteSpaceAtBeginOfParagraph
		// EmptyLineRule commented out
		"TOO_LONG_SENTENCE",               // LongSentenceRule
		"TOO_LONG_PARAGRAPH",              // LongParagraphRule
		"PARAGRAPH_REPEAT_BEGINNING_RULE", // ParagraphRepeatBeginningRule
		"FILLER_WORDS_RU",                 // RussianFillerWordsRule
		// PunctuationMarkAtParagraphEnd commented out
		"PUNCTUATION_PARAGRAPH_END2", // PunctuationMarkAtParagraphEnd2
		// ReadabilityRule pair commented out
		// specific to Russian:
		"MORFOLOGIK_RULE_RU_RU_YO", // MorfologikRussianYOSpellerRule (default off)
		"RU_UNPAIRED_BRACKETS",     // RussianUnpairedBracketsRule
		"RU_COMPOUNDS",             // RussianCompoundRule
		"RU_SIMPLE_REPLACE",        // RussianSimpleReplaceRule
		"WORD_REPEAT_RULE",         // RussianSimpleWordRepeatRule
		"RU_WORD_COHERENCY",        // RussianWordCoherencyRule
		"RU_WORD_REPEAT",           // RussianWordRepeatRule
		"RU_WORD_ROOT_REPEAT",      // RussianWordRootRepeatRule
		"RU_VERB_CONJUGATION",      // RussianVerbConjugationRule
		"RU_DASH_RULE",             // RussianDashRule
		"RU_SPECIFIC_CASE",         // RussianSpecificCaseRule
	}, ids)
	require.Equal(t, ids, Russian.GetRelevantRuleIDs())
	require.Equal(t, ids, NewRussian().GetRelevantRuleIDs())
	// invent / SharedLayout extras / commented-out rules must not reappear
	for _, bad := range []string{
		"DOUBLE_PUNCTUATION", "EMPTY_LINE", "UNPAIRED_BRACKETS",
		"WHITESPACE_PUNCTUATION", "PUNCTUATION_PARAGRAPH_END",
		"READABILITY_RULE_SIMPLE", "READABILITY_RULE_DIFFICULT",
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE",
		"MORFOLOGIK_RULE_RU", "RU_WORD_REPEAT_BEGINNING_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Russian.getPriorityForId — switch keys as Java writes them
// (MORFOLOGIC_* / RUSSIAN_SIMPLE_REPLACE_RULE even when getId differs).
func TestRussianPriorityForId_OnLang(t *testing.T) {
	ru := NewRussian()
	require.Equal(t, 12, ru.GetPriorityForId("RU_DASH_RULE"))
	require.Equal(t, 11, ru.GetPriorityForId("RU_COMPOUNDS"))
	require.Equal(t, 10, ru.GetPriorityForId("RUSSIAN_SIMPLE_REPLACE_RULE"))
	require.Equal(t, 9, ru.GetPriorityForId("RUSSIAN_SPECIFIC_CASE"))
	require.Equal(t, 2, ru.GetPriorityForId("MORFOLOGIC_RULE_RU_RU_YO"))
	require.Equal(t, 1, ru.GetPriorityForId("MORFOLOGIC_RULE_RU_RU"))
	require.Equal(t, -1, ru.GetPriorityForId("Word_root_repeat"))
	require.Equal(t, -2, ru.GetPriorityForId("PUNCT_DPT_2"))
	require.Equal(t, -15, ru.GetPriorityForId("TOO_LONG_PARAGRAPH"))
	// Language base: STYLE category → -50
	require.Equal(t, -50, ru.GetPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, ru.GetPriorityForId("UNKNOWN_RU_XYZ"))
	// getId-spelling variants that are NOT in the Java switch fall through
	require.Equal(t, 0, ru.GetPriorityForId("MORFOLOGIK_RULE_RU_RU"))
	require.Equal(t, 0, ru.GetPriorityForId("RU_SIMPLE_REPLACE"))
}

// Twin of Russian.getRelevantLanguageModelRules.
func TestRussianLanguageModelRuleIDs_OnLang(t *testing.T) {
	ru := NewRussian()
	require.Equal(t, []string{"CONFUSION_RULE"}, ru.GetRelevantLanguageModelRuleIDs())
	require.Equal(t, RussianLanguageModelRelevantRuleIDs(), ru.GetRelevantLanguageModelRuleIDs())
}

// Twin of Russian createDefault* factory wiring → real accepted RU twins.
func TestRussianCreateDefaultFactories(t *testing.T) {
	ru := NewRussian()

	// createDefaultTagger → RussianTagger.INSTANCE over official russian.dict
	require.NotEmpty(t, rutag.DiscoverRussianPOSDict(),
		"official russian.dict must be discoverable")
	tagger := ru.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &rutag.RussianTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	require.Same(t, rutag.DefaultRussianTagger, tagger)
	// Spot-check a known tagged form (same surface RussianTaggerTest uses)
	readings := tagger.Tag([]string{"дом"})
	require.NotEmpty(t, readings)
	require.True(t, readings[0].IsTagged(), "RussianTagger must tag 'дом' from russian.dict")

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("ru")
	st := ru.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Это первое предложение. А это второе.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → RussianWordTokenizer
	wt := ru.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &rutok.RussianWordTokenizer{}, wt)
	// б/у stays whole (Java RussianWordTokenizer)
	toks := wt.Tokenize("купить б/у телефон")
	require.Contains(t, toks, "б/у")
	require.Contains(t, toks, "купить")

	// createDefaultDisambiguator → RussianHybridDisambiguator.getInstance()
	dis := ru.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &rudis.RussianHybridDisambiguator{}, dis)
	require.Same(t, rudis.GetInstance(), dis)
	// Stages wired from official multiwords + disambiguation.xml when present
	require.NotNil(t, dis.Chunker, "MultiWordChunker(/ru/multiwords.txt) must load")
	require.NotNil(t, dis.Rules, "XmlRuleDisambiguator(Russian) must load")

	// createDefaultPostDisambiguationChunker → new RussianChunker()
	// (pre-disambig Chunker remains null — not asserted here)
	pc := ru.CreateDefaultPostDisambiguationChunker()
	require.NotNil(t, pc)
	require.IsType(t, &chunking.RussianChunker{}, pc)

	// createDefaultSynthesizer → RussianSynthesizer.INSTANCE
	require.NotEmpty(t, rusynth.DiscoverRussianSynthDict(),
		"official russian_synth.dict must be discoverable")
	synth := ru.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &rusynth.RussianSynthesizer{}, synth)
	require.Same(t, rusynth.Instance(), synth)
	require.Same(t, rusynth.INSTANCE, synth)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_RU_RU
	require.Equal(t, "MORFOLOGIK_RULE_RU_RU", ru.GetDefaultSpellingRuleID())
}
