package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rosynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ro"
	rodis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ro"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	rotag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ro"
	rotok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ro"
	"github.com/stretchr/testify/require"
)

// Twin of Romanian metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
func TestRomanianMetadata(t *testing.T) {
	ro := NewRomanian()
	require.Equal(t, "ro", ro.GetShortCode())
	require.Equal(t, "ro", RomanianShortCode())
	require.Equal(t, "Romanian", ro.GetName())
	require.Equal(t, []string{"RO"}, ro.GetCountries())
	// Java len(countries)==1 → "ro-RO"
	require.Equal(t, "ro-RO", ro.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "ro/common_words.txt", ro.GetCommonWordsPath())
	// Java Romanian does not override getMaintainedState → LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, ro.GetMaintainedState())
	require.Equal(t, "Ionuț Păduraru", ro.GetMaintainers()[0].Name)
	require.Equal(t, "http://www.archeus.ro", ro.GetMaintainers()[0].URL)
	require.Equal(t, "MORFOLOGIK_RULE_RO_RO", ro.GetDefaultSpellingRuleID())
	require.Equal(t, ro, Romanian)
}

// Twin of Romanian.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestRomanianRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := RomanianRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"MORFOLOGIK_RULE_RO_RO",        // MorfologikRomanianSpellerRule
		"ROMANIAN_WORD_REPEAT_BEGINNING_RULE",
		"RO_SIMPLE_REPLACE", // SimpleReplaceRule
		"RO_COMPOUND",       // CompoundRule
	}, ids)
	require.Equal(t, ids, Romanian.GetRelevantRuleIDs())
	require.Equal(t, ids, NewRomanian().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong IDs must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"MORFOLOGIK_RULE_RO", "RO_UNPAIRED_BRACKETS",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Romanian createDefault* factory wiring → real accepted RO twins.
func TestRomanianCreateDefaultFactories(t *testing.T) {
	ro := NewRomanian()

	// createDefaultTagger → RomanianTagger over official romanian.dict
	require.NotEmpty(t, rotag.DiscoverRomanianPOSDict(),
		"official romanian.dict must be discoverable")
	tagger := ro.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &rotag.RomanianTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	// Spot-check a known tagged form (same surface RomanianTaggerTest uses)
	readings := tagger.Tag([]string{"sunt"})
	require.NotEmpty(t, readings)
	require.True(t, readings[0].IsTagged(), "RomanianTagger must tag 'sunt' from romanian.dict")

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("ro")
	st := ro.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Aceasta este o propoziție. Iar aceasta este a doua.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → RomanianWordTokenizer
	wt := ro.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &rotok.RomanianWordTokenizer{}, wt)
	toks := wt.Tokenize("Aceasta este un test.")
	require.Contains(t, toks, "Aceasta")
	require.Contains(t, toks, "test")

	// createDefaultDisambiguator → XmlRuleDisambiguator(Romanian)
	require.NotEmpty(t, rodis.DiscoverRomanianDisambiguationXML(),
		"official ro/disambiguation.xml must be discoverable")
	dis := ro.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &disambigrules.XmlRuleDisambiguator{}, dis)
	require.NotEmpty(t, dis.Rules, "XmlRuleDisambiguator(Romanian) must load rules")
	require.Same(t, rodis.RomanianXmlRuleDisambiguator(), dis)

	// createDefaultSynthesizer → RomanianSynthesizer.INSTANCE
	require.NotEmpty(t, rosynth.DiscoverRomanianSynthDict(),
		"official romanian_synth.dict must be discoverable")
	synth := ro.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &rosynth.RomanianSynthesizer{}, synth)
	require.Same(t, rosynth.Instance(), synth)
	require.Same(t, rosynth.INSTANCE, synth)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_RO_RO
	require.Equal(t, "MORFOLOGIK_RULE_RO_RO", ro.GetDefaultSpellingRuleID())
}
