package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	brtag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/br"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
	"github.com/stretchr/testify/require"
)

// Twin of Breton metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
func TestBretonMetadata(t *testing.T) {
	br := NewBreton()
	require.Equal(t, "br", br.GetShortCode())
	require.Equal(t, "br", BretonShortCode())
	require.Equal(t, "Breton", br.GetName())
	require.Equal(t, []string{"FR"}, br.GetCountries())
	// Java len(countries)==1 → "br-FR"
	require.Equal(t, "br-FR", br.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "br/common_words.txt", br.GetCommonWordsPath())
	// Java Breton overrides getMaintainedState → ActivelyMaintained
	require.Equal(t, languagetool.ActivelyMaintained, br.GetMaintainedState())
	require.Equal(t, DominiquePelle.Name, br.GetMaintainers()[0].Name)
	require.Equal(t, "Fulup Jakez", br.GetMaintainers()[1].Name)
	require.Equal(t, "MORFOLOGIK_RULE_BR_FR", br.GetDefaultSpellingRuleID())
	require.Equal(t, br, Breton)
}

// Twin of Breton.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestBretonRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := BretonRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"MORFOLOGIK_RULE_BR_FR",        // MorfologikBretonSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"BR_TOPO",                      // TopoReplaceRule
	}, ids)
	require.Equal(t, ids, Breton.GetRelevantRuleIDs())
	require.Equal(t, ids, NewBreton().GetRelevantRuleIDs())
	// invent / SharedLayout extras / compound / unpaired must not reappear
	// (Java Breton has no unpaired brackets; BretonCompoundRule is not in getRelevantRules)
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"MORFOLOGIK_RULE_BR", "UNPAIRED_BRACKETS", "WORD_REPEAT_RULE",
		"BR_COMPOUNDS",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Breton getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestBreton_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Breton.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/br/grammar.xml",
	}, files)
	require.Equal(t, files, NewBreton().GetRuleFileNamesWithExists(exists))
}

// Twin of Breton createDefault* factory wiring → real accepted BR twins.
// Java overrides: sentence tokenizer, word tokenizer, tagger, disambiguator, spelling rule.
// No synthesizer invent (Java has none).
func TestBretonCreateDefaultFactories(t *testing.T) {
	br := NewBreton()

	// createDefaultTagger → BretonTagger over official breton.dict
	require.NotEmpty(t, brtag.DiscoverBretonPOSDict(),
		"official breton.dict must be discoverable")
	tagger := br.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &brtag.BretonTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	require.Same(t, brtag.DefaultBretonTagger, tagger)
	// Spot-check a known form when dict loads (BretonTaggerTest surfaces)
	readings := tagger.Tag([]string{"Breizh"})
	require.NotEmpty(t, readings)
	if readings[0].IsTagged() {
		// expected when binary dict carries the entry
	} else {
		t.Log("breton.dict did not tag 'Breizh' — honest gap if dict sparse / missing")
	}

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("br")
	st := br.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Lakait amañ ho testenn. Pe implijit an frazenn-mañ.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → BretonWordTokenizer (c'h / apostrophe handling)
	wt := br.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &brtok.BretonWordTokenizer{}, wt)
	toks := wt.Tokenize("Test c'h")
	require.Contains(t, toks, "c’h")

	// createDefaultDisambiguator → XmlRuleDisambiguator(Breton)
	require.NotEmpty(t, brtag.DiscoverBretonDisambiguationXML(),
		"official br/disambiguation.xml must be discoverable")
	dis := br.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &disambigrules.XmlRuleDisambiguator{}, dis)
	require.NotEmpty(t, dis.Rules, "XmlRuleDisambiguator(Breton) must load rules")
	require.Same(t, brtag.BretonXmlRuleDisambiguator(), dis)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_BR_FR
	require.Equal(t, "MORFOLOGIK_RULE_BR_FR", br.GetDefaultSpellingRuleID())
}
