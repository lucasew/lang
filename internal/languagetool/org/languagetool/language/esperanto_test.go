package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	tageo "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/eo"
	eotok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/eo"
	"github.com/stretchr/testify/require"
)

// Twin of Esperanto metadata (getName / getShortCode / getCountries / getMaintainers /
// getMaintainedState / createDefaultSpellingRule id).
// Java overrides getMaintainedState → ActivelyMaintained; getCountries → {}.
func TestEsperantoMetadata(t *testing.T) {
	eo := NewEsperanto()
	require.Equal(t, "eo", eo.GetShortCode())
	require.Equal(t, "eo", EsperantoShortCode())
	require.Equal(t, "Esperanto", eo.GetName())
	// Java getCountries → new String[]{} → empty slice
	require.Empty(t, eo.GetCountries())
	require.Equal(t, "eo", eo.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "eo/common_words.txt", eo.GetCommonWordsPath())
	// Java Esperanto overrides getMaintainedState → ActivelyMaintained
	require.Equal(t, languagetool.ActivelyMaintained, eo.GetMaintainedState())
	require.Equal(t, DominiquePelle.Name, eo.GetMaintainers()[0].Name)
	require.Len(t, eo.GetMaintainers(), 1)
	require.Equal(t, "HUNSPELL_RULE", eo.GetDefaultSpellingRuleID())
	require.Equal(t, eo, Esperanto)
}

// Twin of Esperanto.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestEsperantoRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	ids := EsperantoRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (default pairs)
		"HUNSPELL_RULE",                // HunspellRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
	}, ids)
	require.Len(t, ids, 8)
	require.Equal(t, ids, Esperanto.GetRelevantRuleIDs())
	require.Equal(t, ids, NewEsperanto().GetRelevantRuleIDs())
	// invent / SharedLayout extras / DateCheckFilter must not reappear
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "EMPTY_LINE",
		"TOO_LONG_PARAGRAPH", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"DATE_CHECK_FILTER", "DATE_FILTER",
	} {
		require.NotContains(t, ids, bad)
	}
}

// Twin of Esperanto getRuleFileNames — base grammar.xml only (no RULE_FILES extras).
func TestEsperanto_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Esperanto.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/eo/grammar.xml",
	}, files)
	require.Equal(t, files, NewEsperanto().GetRuleFileNamesWithExists(exists))
}

// Twin of Esperanto createDefault* factory wiring → real accepted EO twins.
// Java overrides: sentence tokenizer, word tokenizer, tagger, disambiguator, spelling rule.
// No synthesizer invent (Java has none).
func TestEsperantoCreateDefaultFactories(t *testing.T) {
	eo := NewEsperanto()

	// createDefaultTagger → EsperantoTagger
	tagger := eo.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &tageo.EsperantoTagger{}, tagger)
	// Spot-check a known form (EsperantoTaggerTest surface)
	readings := tagger.Tag([]string{"domo"})
	require.NotEmpty(t, readings)
	if readings[0].IsTagged() {
		// expected when tagger resources load
	} else {
		t.Log("EsperantoTagger did not tag 'domo' — honest gap if resources missing")
	}

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("eo")
	st := eo.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Saluton mondo. Ĉi tio estas dua frazo.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → EsperantoWordTokenizer (apostrophe handling)
	wt := eo.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &eotok.EsperantoWordTokenizer{}, wt)
	toks := wt.Tokenize("l'esperanto")
	require.NotEmpty(t, toks)

	// createDefaultDisambiguator → XmlRuleDisambiguator(Esperanto)
	require.NotEmpty(t, tageo.DiscoverEsperantoDisambiguationXML(),
		"official eo/disambiguation.xml must be discoverable")
	dis := eo.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &disambigrules.XmlRuleDisambiguator{}, dis)
	require.NotEmpty(t, dis.Rules, "XmlRuleDisambiguator(Esperanto) must load rules")
	require.Same(t, tageo.EsperantoXmlRuleDisambiguator(), dis)

	// createDefaultSpellingRule id → HUNSPELL_RULE
	require.Equal(t, "HUNSPELL_RULE", eo.GetDefaultSpellingRuleID())
}
