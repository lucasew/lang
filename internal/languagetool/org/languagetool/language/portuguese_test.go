package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	ptsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
	ptdis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pt"
	pttag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pt"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
	"github.com/stretchr/testify/require"
)

// Twin of Portuguese / PortugalPortuguese metadata (default product surface).
func TestPortugueseMetadata(t *testing.T) {
	pt := NewPortuguese()
	require.Equal(t, "pt", pt.GetShortCode())
	require.Equal(t, "pt", PortugueseShortCode())
	// Product default surface is PortugalPortuguese (Java getDefaultLanguageVariant → pt-PT).
	require.Equal(t, "Portuguese (Portugal)", pt.GetName())
	require.Equal(t, []string{"PT"}, pt.GetCountries())
	require.Equal(t, "pt-PT", pt.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "pt-PT", pt.GetDefaultLanguageVariantCode())
	require.Equal(t, "pt-PT", PortugueseDefaultLanguageVariantCode())
	require.Equal(t, "pt/common_words.txt", pt.GetCommonWordsPath())
	require.Equal(t, languagetool.ActivelyMaintained, pt.GetMaintainedState())
	require.Equal(t, "Marco A.G. Pinto", pt.GetMaintainers()[0].Name)
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT", pt.GetDefaultSpellingRuleID())
	require.Equal(t, pt, Portuguese)
	require.Equal(t, PortugalPortuguese, Portuguese)

	// Base Portuguese.getCountries (Java class Portuguese, not locale variants).
	require.Equal(t, []string{"", "CV", "GW", "MO", "ST", "TL"}, PortugueseBaseCountries())

	// Quotes: PortugalPortuguese overrides double quotes to « ».
	require.Equal(t, "«", pt.GetOpeningDoubleQuote())
	require.Equal(t, "»", pt.GetClosingDoubleQuote())
	require.Equal(t, "‘", pt.GetOpeningSingleQuote())
	require.Equal(t, "’", pt.GetClosingSingleQuote())
	require.True(t, pt.IsAdvancedTypographyEnabled())

	// Brazilian surface uses base Portuguese curly doubles.
	br := NewBrazilianPortuguese()
	require.Equal(t, "Portuguese (Brazil)", br.GetName())
	require.Equal(t, []string{"BR"}, br.GetCountries())
	require.Equal(t, "pt-BR", br.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "“", br.GetOpeningDoubleQuote())
	require.Equal(t, "”", br.GetClosingDoubleQuote())
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", br.GetDefaultSpellingRuleID())
}

// Twin of Portuguese.getRelevantRules — class getId order exactly as Java Arrays.asList.
func TestPortugueseRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	base := PortugueseRelevantRuleIDs()
	require.Equal(t, []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_PT_PT",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"EMPTY_LINE",
		"PARAGRAPH_REPEAT_BEGINNING_RULE",
		"PUNCTUATION_PARAGRAPH_END",
		"PT_COMPOUNDS_POST_REFORM",
		"PT_COLOUR_HYPHENATION",
		"PT_SIMPLE_REPLACE_ORTHOGRAPHY",
		"PT_SIMPLE_REPLACE",
		"PT_BARBARISMS_REPLACE",
		// PortugueseArchaismsRule commented out in Java base (#3095)
		"PT_CLICHE_REPLACE",
		"FILLER_WORDS_PT",
		"PT_REDUNDANCY_REPLACE",
		"PT_WORDINESS_REPLACE",
		// PortugueseWeaselWordsRule commented out in Java
		"PT_WIKIPEDIA_COMMON_ERRORS",
		"PORTUGUESE_WORD_REPEAT_RULE",
		"PORTUGUESE_WORD_REPEAT_BEGINNING_RULE",
		"ACCENTUATION_CHECK_PT",
		"PT_DIACRITICS_REPLACE",
		"PORTUGUESE_WRONG_WORD_IN_CONTEXT",
		"PT_WORD_COHERENCY",
		"UNIDADES_METRICAS",
		"READABILITY_RULE_SIMPLE_PT",
		"READABILITY_RULE_DIFFICULT_PT",
		"DOUBLE_PUNCTUATION",
		"PT_ENGLISH_CONTRACTION_ORTHOGRAPHY",
	}, base)

	// invent / wrong IDs must not reappear on base list
	for _, bad := range []string{
		"MULTIPLE_WHITESPACE", "COMMA_WHITESPACE", "PT_UNPAIRED_BRACKETS",
		"PT_ARCHAISMS_REPLACE", "PT_WEASELWORD_REPLACE",
		"WHITESPACE_PUNCTUATION", "MORFOLOGIK_RULE_PT",
		"PT_POSAO_DASH_RULE", "PT_AGREEMENT_REPLACE", "PT_PT_SIMPLE_REPLACE",
	} {
		require.NotContains(t, base, bad)
	}

	// PortugalPortuguese = base + extras (Java ArrayList after super)
	ids := PortugalPortugueseRelevantRuleIDs()
	require.Equal(t, append(append([]string(nil), base...), PortugalPortugueseRelevantExtraRuleIDs()...), ids)
	require.Equal(t, []string{
		"PT_COMPOUNDS_POST_REFORM",
		"PT_POSAO_DASH_RULE",
		"PT_AGREEMENT_REPLACE",
		"PT_PT_SIMPLE_REPLACE",
		"PT_BARBARISMS_REPLACE",
		"PT_ARCHAISMS_REPLACE",
		"PT_CLICHE_REPLACE",
		"PT_REDUNDANCY_REPLACE",
		"PT_WORDINESS_REPLACE",
		"PT_WIKIPEDIA_COMMON_ERRORS",
	}, PortugalPortugueseRelevantExtraRuleIDs())
	require.Equal(t, ids, Portuguese.GetRelevantRuleIDs())
	require.Equal(t, ids, NewPortuguese().GetRelevantRuleIDs())
	require.Equal(t, ids, NewPortugalPortuguese().GetRelevantRuleIDs())
	// Extras appear after base tail
	require.Equal(t, "PT_ENGLISH_CONTRACTION_ORTHOGRAPHY", ids[len(base)-1])
	require.Equal(t, "PT_COMPOUNDS_POST_REFORM", ids[len(base)])
	require.Equal(t, "PT_WIKIPEDIA_COMMON_ERRORS", ids[len(ids)-1])

	br := NewBrazilianPortuguese().GetRelevantRuleIDs()
	require.Contains(t, br, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, br, "MORFOLOGIK_RULE_PT_PT")
	// BR sector not yet wiring its extras — base only
	require.NotContains(t, br, "PT_PT_SIMPLE_REPLACE")
	require.NotContains(t, br, "PT_AGREEMENT_REPLACE")
}

// Twin of Portuguese.getPriorityForId / PortugalPortuguese override.
func TestPortuguesePriorityForId_OnLang(t *testing.T) {
	pt := NewPortugalPortuguese()
	// PortugalPortuguese map first
	require.Equal(t, 1, pt.GetPriorityForId("PT_COMPOUNDS_POST_REFORM"))
	require.Equal(t, -9, pt.GetPriorityForId("PORTUGUESE_OLD_SPELLING_INTERNAL"))
	// base Portuguese map
	require.Equal(t, 5, pt.GetPriorityForId("HOMOPHONE_AS_CARD"))
	require.Equal(t, -997, pt.GetPriorityForId("TOO_LONG_SENTENCE"))
	require.Equal(t, -998, pt.GetPriorityForId("TOO_LONG_PARAGRAPH"))
	// prefix gates
	require.Equal(t, -50, pt.GetPriorityForId("MORFOLOGIK_RULE_PT_PT"))
	// Language base: STYLE category → -50
	require.Equal(t, -50, pt.GetPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, pt.GetPriorityForId("UNKNOWN_PT_XYZ"))

	// pt-BR uses base Portuguese only (no Portugal map override for compounds)
	br := NewBrazilianPortuguese()
	require.Equal(t, -45, br.GetPriorityForId("PT_COMPOUNDS_POST_REFORM"))
	require.Equal(t, -50, br.GetPriorityForId("MORFOLOGIK_RULE_PT_BR"))
}

// Twin of Portuguese createDefault* factory wiring → real accepted PT twins.
func TestPortugueseCreateDefaultFactories(t *testing.T) {
	pt := NewPortuguese()

	// createDefaultTagger → PortugueseTagger over official portuguese.dict
	require.NotEmpty(t, pttag.DiscoverPortuguesePOSDict(),
		"official portuguese.dict must be discoverable")
	tagger := pt.CreateDefaultTagger()
	require.NotNil(t, tagger)
	require.IsType(t, &pttag.PortugueseTagger{}, tagger)
	require.NotNil(t, tagger.GetWordTagger())
	// Spot-check a known tagged form
	readings := tagger.Tag([]string{"casa"})
	require.NotEmpty(t, readings)
	require.True(t, readings[0].IsTagged(), "PortugueseTagger must tag 'casa' from portuguese.dict")

	// createDefaultSentenceTokenizer → SRXSentenceTokenizer("pt")
	st := pt.CreateDefaultSentenceTokenizer()
	require.NotNil(t, st)
	sents := st.Tokenize("Isto é uma frase. E isto é outra.")
	require.GreaterOrEqual(t, len(sents), 2)

	// createDefaultWordTokenizer → PortugueseWordTokenizer (+ Java ctor tagger)
	wt := pt.CreateDefaultWordTokenizer()
	require.NotNil(t, wt)
	require.IsType(t, &pttok.PortugueseWordTokenizer{}, wt)
	toks := wt.Tokenize("Isto é uma frase.")
	require.Contains(t, toks, "Isto")
	require.Contains(t, toks, "frase")

	// createDefaultDisambiguator → PortugueseHybridDisambiguator
	dis := pt.CreateDefaultDisambiguator()
	require.NotNil(t, dis)
	require.IsType(t, &ptdis.PortugueseHybridDisambiguator{}, dis)
	// Stages wired from official multiwords + spelling_global + disambiguation.xml when present
	require.NotNil(t, dis.GlobalChunker, "MultiWordChunker(/spelling_global.txt) must load")
	require.NotNil(t, dis.Chunker, "MultiWordChunker(/pt/multiwords.txt) must load")
	require.NotNil(t, dis.Rules, "XmlRuleDisambiguator(Portuguese) must load")

	// createDefaultSynthesizer → PortugueseSynthesizer.INSTANCE
	require.NotEmpty(t, ptsynth.DiscoverPortugueseSynthDict(),
		"official portuguese_synth.dict must be discoverable")
	synth := pt.CreateDefaultSynthesizer()
	require.NotNil(t, synth)
	require.IsType(t, &ptsynth.PortugueseSynthesizer{}, synth)
	require.Same(t, ptsynth.Instance(), synth)
	require.Same(t, ptsynth.INSTANCE, synth)

	// createDefaultSpellingRule id → MORFOLOGIK_RULE_PT_PT
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT", pt.GetDefaultSpellingRuleID())

	// Language-model relevant rules
	require.Equal(t, []string{"CONFUSION_RULE"}, pt.GetRelevantLanguageModelRuleIDs())
}
