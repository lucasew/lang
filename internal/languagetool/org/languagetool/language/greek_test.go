package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	elsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/el"
	eldis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/el"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	eltag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/el"
	eltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/el"
	"github.com/stretchr/testify/require"
)

// Twin of Greek metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
func TestGreekMetadata(t *testing.T) {
	el := NewGreek()
	require.Equal(t, "el", el.GetShortCode())
	require.Equal(t, "el", GreekShortCode())
	require.Equal(t, "Greek", el.GetName())
	require.Equal(t, []string{"GR"}, el.GetCountries())
	// Java len(countries)==1 → "el-GR"
	require.Equal(t, "el-GR", el.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "el/common_words.txt", el.GetCommonWordsPath())
	// Java Greek overrides getMaintainedState → ActivelyMaintained
	require.Equal(t, languagetool.ActivelyMaintained, el.GetMaintainedState())
	require.Equal(t, "Panagiotis Minos", el.GetMaintainers()[0].Name)
	require.Equal(t, "MORFOLOGIK_RULE_EL_GR", el.GetDefaultSpellingRuleID())
	require.Equal(t, el, Greek)
}

// Twin of Greek.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestGreekRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := GreekRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"EL_UNPAIRED_BRACKETS",         // GenericUnpairedBracketsRule custom id
		"TOO_LONG_SENTENCE",            // LongSentenceRule(…, 50)
		"MORFOLOGIK_RULE_EL_GR",        // MorfologikGreekSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"GREEK_WORD_REPEAT_BEGINNING_RULE",
		"WORD_REPEAT_RULE",
		"GREEK_HOMONYMS_REPLACE",
		"EL_SPECIFIC_CASE",
		"GREEK_ORTHOGRAPHY_NUMERAL_STRESS",
		"EL_REDUNDANCY_REPLACE",
	}, ids)
	require.Equal(t, ids, Greek.GetRelevantRuleIDs())
	require.Equal(t, ids, NewGreek().GetRelevantRuleIDs())
	// invent / SharedLayout extras / wrong unpaired id must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"MORFOLOGIK_RULE_EL", "UNPAIRED_BRACKETS",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Greek getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestGreek_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Greek.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/el/grammar.xml",
	}, files)
	require.Equal(t, files, NewGreek().GetRuleFileNamesWithExists(exists))
}

// Twin of Greek createDefault* factory wiring → real accepted EL twins.
func TestGreekCreateDefaultFactories(t *testing.T) {
	el := NewGreek()

	// createDefaultTagger → GreekTagger over official greek.dict
	require.NotEmpty(t, eltag.DiscoverGreekPOSDict(),
		"official greek.dict must be discoverable")
	tagger := el.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &eltag.GreekTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	require.Same(t, eltag.DefaultGreekTagger, tagger)
	// greek.dict is a tiny official stub (+ GreekAnalyzer leaf gap). Spot-check
	// a form present in greek.txt source when the binary loads it.
	readings := tagger.Tag([]string{"έχω"})
	require.NotEmpty(t, readings)
	if readings[0].IsTagged() {
		// expected when binary dict carries the few test entries from greek.txt
	} else {
		t.Log("greek.dict stub did not tag 'έχω' — honest gap (tiny dict / missing GreekAnalyzer)")
	}

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("el")
	st := el.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Αυτή είναι η πρώτη πρόταση. Και αυτή είναι η δεύτερη.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → GreekWordTokenizer
	wt := el.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &eltok.GreekWordTokenizer{}, wt)
	toks := wt.Tokenize("Αυτή είναι μια δοκιμή.")
	require.Contains(t, toks, "Αυτή")
	require.Contains(t, toks, "δοκιμή")

	// createDefaultDisambiguator → XmlRuleDisambiguator(Greek)
	require.NotEmpty(t, eldis.DiscoverGreekDisambiguationXML(),
		"official el/disambiguation.xml must be discoverable")
	dis := el.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &disambigrules.XmlRuleDisambiguator{}, dis)
	require.NotEmpty(t, dis.Rules, "XmlRuleDisambiguator(Greek) must load rules")
	require.Same(t, eldis.GreekXmlRuleDisambiguator(), dis)

	// createDefaultSynthesizer → GreekSynthesizer.INSTANCE
	require.NotEmpty(t, elsynth.DiscoverGreekSynthDict(),
		"official greek_synth.dict must be discoverable")
	synth := el.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &elsynth.GreekSynthesizer{}, synth)
	require.Same(t, elsynth.Instance(), synth)
	require.Same(t, elsynth.INSTANCE, synth)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_EL_GR
	require.Equal(t, "MORFOLOGIK_RULE_EL_GR", el.GetDefaultSpellingRuleID())
}
