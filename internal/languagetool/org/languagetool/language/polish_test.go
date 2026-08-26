package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	plsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pl"
	pldis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pl"
	pltag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	pltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pl"
	"github.com/stretchr/testify/require"
)

// Twin of Polish metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
func TestPolishMetadata(t *testing.T) {
	pl := NewPolish()
	require.Equal(t, "pl", pl.GetShortCode())
	require.Equal(t, "pl", PolishShortCode())
	require.Equal(t, "Polish", pl.GetName())
	require.Equal(t, []string{"PL"}, pl.GetCountries())
	// Java len(countries)==1 → "pl-PL"
	require.Equal(t, "pl-PL", pl.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "pl/common_words.txt", pl.GetCommonWordsPath())
	require.Equal(t, languagetool.ActivelyMaintained, pl.GetMaintainedState())
	require.Equal(t, MarcinMilkowski.Name, pl.GetMaintainers()[0].Name)
	require.Equal(t, "http://marcinmilkowski.pl", pl.GetMaintainers()[0].URL)
	require.Equal(t, "MORFOLOGIK_RULE_PL_PL", pl.GetDefaultSpellingRuleID())
	require.Equal(t, pl, Polish)
}

// Twin of Polish.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestPolishRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := PolishRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"PL_UNPAIRED_BRACKETS",         // PolishUnpairedBracketsRule
		"MORFOLOGIK_RULE_PL_PL",        // MorfologikPolishSpellerRule
		"PL_WORD_REPEAT",               // PolishWordRepeatRule
		"PL_COMPOUNDS",                 // CompoundRule
		"PL_SIMPLE_REPLACE",            // SimpleReplaceRule
		"PL_WORD_COHERENCY",            // WordCoherencyRule
		"DASH_RULE",                    // DashRule / AbstractDashRule
	}, ids)
	require.Equal(t, ids, Polish.GetRelevantRuleIDs())
	require.Equal(t, ids, NewPolish().GetRelevantRuleIDs())
	// invent / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "UNPAIRED_BRACKETS",
		"PL_SENTENCE_WHITESPACE", "MORFOLOGIK_RULE_PL", "PL_DASH_RULE",
		"DOUBLE_PUNCTUATION", "EMPTY_LINE", "TOO_LONG_PARAGRAPH",
		"PL_WORD_REPEAT_BEGINNING_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Polish.getPriorityForId — ZDANIA_ZLOZONE = -1; else Language base.
func TestPolishPriorityForId_OnLang(t *testing.T) {
	pl := NewPolish()
	require.Equal(t, -1, pl.GetPriorityForId("ZDANIA_ZLOZONE"))
	require.Equal(t, -1, PolishPriorityForId("ZDANIA_ZLOZONE"))
	// Language base: STYLE category → -50
	require.Equal(t, -50, pl.GetPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, pl.GetPriorityForId("UNKNOWN_PL_XYZ"))
}

// Twin of Polish createDefault* factory wiring → real accepted PL twins.
func TestPolishCreateDefaultFactories(t *testing.T) {
	pl := NewPolish()

	// createDefaultTagger → PolishTagger over official polish.dict
	require.NotEmpty(t, pltag.DiscoverPolishPOSDict(),
		"official polish.dict must be discoverable")
	tagger := pl.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &pltag.PolishTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	// Spot-check a known tagged form (same surface PolishTaggerTest uses)
	readings := tagger.Tag([]string{"dom"})
	require.NotEmpty(t, readings)
	require.True(t, readings[0].IsTagged(), "PolishTagger must tag 'dom' from polish.dict")

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("pl")
	st := pl.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("To jest zdanie. A to drugie.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → PolishWordTokenizer + setTagger(getTagger())
	wt := pl.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &pltok.PolishWordTokenizer{}, wt)
	// With tagger, hybrid hyphen split for adja compounds (Java PolishWordTokenizerTest)
	compound := wt.Tokenize("To jest kobieta-wojownik.")
	require.Contains(t, compound, "kobieta")
	require.Contains(t, compound, "-")
	require.Contains(t, compound, "wojownik")

	// createDefaultDisambiguator → PolishHybridDisambiguator
	dis := pl.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &pldis.PolishHybridDisambiguator{}, dis)
	// Stages wired from official multiwords + disambiguation.xml when present
	require.NotNil(t, dis.Chunker, "MultiWordChunker(/pl/multiwords.txt) must load")
	require.NotNil(t, dis.Rules, "XmlRuleDisambiguator(Polish) must load")

	// createDefaultSynthesizer → PolishSynthesizer.INSTANCE
	require.NotEmpty(t, plsynth.DiscoverPolishSynthDict(),
		"official polish_synth.dict must be discoverable")
	synth := pl.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &plsynth.PolishSynthesizer{}, synth)
	require.Same(t, plsynth.Instance(), synth)
	require.Same(t, plsynth.INSTANCE, synth)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_PL_PL
	require.Equal(t, "MORFOLOGIK_RULE_PL_PL", pl.GetDefaultSpellingRuleID())
}
