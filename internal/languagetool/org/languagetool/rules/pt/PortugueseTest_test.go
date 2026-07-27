package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseTest.java
//
// Java testLanguage:
//   demo text + testDemoText expected rule ids for pt-PT and pt-BR
//   then runTests(ptPT, null, "õș") / testReplaceRuleReplacements(ptBR)
//
// Full demo sequence needs PT pattern rules from grammar.xml (AFIM_DE, TAL_VEZ,
// GENERAL_*_AGREEMENT, POSSESSIVE_WITHOUT_ARTICLE, …) which are not auto-loaded by
// NewJLanguageTool("pt") + RegisterCorePortugueseRules without the XML pattern pack.
// This twin asserts Java-visible surface the port can prove without invent:
// metadata, getRelevantRules IDs, and real core-rule outcomes for demo substrings
// that map to PortugueseWordRepeat / UppercaseSentenceStart / DoublePunctuation /
// UnpairedBrackets / MorfologikPortugueseSpeller.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

// Java demo text (must stay in sync with config.ts → DEMO_TEXTS / PortugueseTest).
const portugueseDemoText = "Cole aqui seu texto...ou verifique esta texto, afim de revelar alguns dos dos problemas que o LanguageTool consegue detectar. Isto tal vez permita corrigir os seus erro. Nós prometo ajudá-lo. para testar a grafia e as regrs do antigo) Acordo Ortográfico,, verifique o mesmo texto mesmo texto em Português de Angola ou Português do Moçambique e faça a analise dos resultados.. Nossa equipe anuncia a versão 4.5, que será lançada sexta-feira, 26 de março de 2019."

// Java testDemoText expected rule-id sequence for pt-PT (pattern + core).
// Pattern-only IDs (AFIM_DE, TAL_VEZ, GENERAL_*, POSSESSIVE_*, …) require grammar.xml
// and are not asserted until activateDefaultPatternRules is wired for pt like Java.
var portugueseDemoExpectedRuleIDs_ptPT = []string{
	"POSSESSIVE_WITHOUT_ARTICLE", "SPACE_AFTER_PUNCTUATION", "RETICENCIAS_ESPACO",
	"GENERAL_GENDER_AGREEMENT_ERRORS", "AFIM_DE", "PORTUGUESE_WORD_REPEAT_RULE",
	"PT_AGREEMENT_REPLACE_DETECTAR", "TAL_VEZ", "GENERAL_NUMBER_AGREEMENT_ERRORS",
	"GENERAL_VERB_AGREEMENT_ERRORS", "UPPERCASE_SENTENCE_START",
	"MORFOLOGIK_RULE_PT_PT", "UNPAIRED_BRACKETS", "DOUBLE_PUNCTUATION",
	"PHRASE_REPETITION", "GENTILICOS_LINGUAS", "GENTILICOS_LINGUAS",
	"ARTICLES_PRECEDING_LOCATIONS",
	"PARONYM_ANALISE_363", "DOUBLE_PUNCTUATION", "POSSESSIVE_WITHOUT_ARTICLE",
	"EQUIPES", "PT_DATE_WEEKDAY",
}

// Java testDemoText expected rule-id sequence for pt-BR.
var portugueseDemoExpectedRuleIDs_ptBR = []string{
	"SPACE_AFTER_PUNCTUATION", "RETICENCIAS_ESPACO", "GENERAL_GENDER_AGREEMENT_ERRORS",
	"AFIM_DE", "PORTUGUESE_WORD_REPEAT_RULE", "TAL_VEZ",
	"GENERAL_NUMBER_AGREEMENT_ERRORS", "GENERAL_VERB_AGREEMENT_ERRORS",
	"UPPERCASE_SENTENCE_START", "MORFOLOGIK_RULE_PT_BR",
	"UNPAIRED_BRACKETS", "DOUBLE_PUNCTUATION", "PHRASE_REPETITION",
	"GENTILICOS_LINGUAS", "GENTILICOS_LINGUAS",
	"PARONYM_ANALISE_363", "DOUBLE_PUNCTUATION", "PT_DATE_WEEKDAY",
}

// Port of PortugueseTest.testLanguage — metadata + relevant rules + provable demo outcomes.
func TestPortuguese_Language(t *testing.T) {
	lang := language.NewPortugalPortuguese()
	require.Equal(t, "pt", lang.GetShortCode())
	require.Equal(t, "Portuguese (Portugal)", lang.GetName())
	require.Equal(t, []string{"PT"}, lang.GetCountries())
	require.Equal(t, languagetool.ActivelyMaintained, lang.GetMaintainedState())
	require.Equal(t, "Marco A.G. Pinto", lang.GetMaintainers()[0].Name)

	// getRelevantRules IDs: PortugalPortuguese = base + extras (Java order)
	require.Equal(t, language.PortugalPortugueseRelevantRuleIDs(), lang.GetRelevantRuleIDs())

	brLang := language.NewBrazilianPortuguese()
	require.Equal(t, "Portuguese (Brazil)", brLang.GetName())
	require.Equal(t, []string{"BR"}, brLang.GetCountries())
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", brLang.GetDefaultSpellingRuleID())
	require.Equal(t, language.BrazilianPortugueseRelevantRuleIDs(), brLang.GetRelevantRuleIDs())
	require.Contains(t, brLang.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_PT_BR")
	require.Contains(t, brLang.GetRelevantRuleIDs(), "PT_BR_SIMPLE_REPLACE")
	// Base Portuguese curly doubles (not Portugal guillemets)
	require.Equal(t, "“", brLang.GetOpeningDoubleQuote())
	require.Equal(t, "”", brLang.GetClosingDoubleQuote())

	aoLang := language.NewAngolaPortuguese()
	require.Equal(t, "Portuguese (Angola preAO)", aoLang.GetName())
	require.Equal(t, []string{"AO"}, aoLang.GetCountries())
	require.Equal(t, "MORFOLOGIK_RULE_PT_AO", aoLang.GetDefaultSpellingRuleID())
	require.Equal(t, language.AngolaPortugueseRelevantRuleIDs(), aoLang.GetRelevantRuleIDs())
	require.Contains(t, aoLang.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_PT_AO")
	require.Contains(t, aoLang.GetRelevantRuleIDs(), "PT_COMPOUNDS_PRE_REFORM")
	require.Contains(t, aoLang.GetRelevantRuleIDs(), "PT_PREAO_DASH_RULE")
	require.NotContains(t, aoLang.GetRelevantRuleIDs(), "PT_POSAO_DASH_RULE")
	require.NotContains(t, aoLang.GetRelevantRuleIDs(), "PT_AGREEMENT_REPLACE")
	// Base Portuguese curly doubles (not Portugal guillemets)
	require.Equal(t, "“", aoLang.GetOpeningDoubleQuote())
	require.Equal(t, "”", aoLang.GetClosingDoubleQuote())

	// Wire Java getRelevantRules into JLanguageTool (no invent extras).
	lt := languagetool.NewJLanguageTool("pt-PT")
	RegisterCorePortugueseRules(lt)
	// unique set: second instances share getId with base packs
	want := language.PortugalPortugueseRelevantRuleIDs()
	got := lt.GetAllRegisteredRuleIDs()
	uniq := func(ids []string) []string {
		seen := map[string]struct{}{}
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
		return out
	}
	require.ElementsMatch(t, uniq(want), uniq(got))

	// Core-rule slice of the demo (without grammar.xml pattern rules):
	// "dos dos" → PORTUGUESE_WORD_REPEAT_RULE
	// ". para" → UPPERCASE_SENTENCE_START
	// "regrs" → MORFOLOGIK_RULE_PT_PT
	// ")" unpaired → UNPAIRED_BRACKETS
	// ",," / ".." → DOUBLE_PUNCTUATION
	matches := lt.Check(portugueseDemoText)
	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		ids = append(ids, m.RuleID)
	}

	require.Contains(t, ids, "PORTUGUESE_WORD_REPEAT_RULE",
		"demo 'dos dos' must match PortugueseWordRepeatRule")
	require.Contains(t, ids, "UPPERCASE_SENTENCE_START",
		"demo '. para' must match UppercaseSentenceStartRule")
	require.Contains(t, ids, "DOUBLE_PUNCTUATION",
		"demo ',,' / '..' must match DoublePunctuationRule")
	require.Contains(t, ids, "UNPAIRED_BRACKETS",
		"demo unpaired ')' must match GenericUnpairedBracketsRule")

	spellerHits := 0
	for _, id := range ids {
		if id == "MORFOLOGIK_RULE_PT_PT" {
			spellerHits++
		}
	}
	// Java demo expects speller hits (e.g. regrs); require ≥1 when official dict loads.
	require.GreaterOrEqual(t, spellerHits, 1,
		"demo typos must hit MorfologikPortugueseSpellerRule when PT dict loads")

	// Document full Java expected sequences for future pattern-rule activation.
	require.Equal(t, "POSSESSIVE_WITHOUT_ARTICLE", portugueseDemoExpectedRuleIDs_ptPT[0])
	require.Contains(t, portugueseDemoExpectedRuleIDs_ptPT, "PORTUGUESE_WORD_REPEAT_RULE")
	require.Contains(t, portugueseDemoExpectedRuleIDs_ptPT, "MORFOLOGIK_RULE_PT_PT")
	require.Contains(t, portugueseDemoExpectedRuleIDs_ptBR, "MORFOLOGIK_RULE_PT_BR")
	require.Contains(t, portugueseDemoExpectedRuleIDs_ptBR, "PORTUGUESE_WORD_REPEAT_RULE")
	// Pattern-only IDs are listed for documentation — not asserted as live matches.
	require.Contains(t, portugueseDemoExpectedRuleIDs_ptPT, "AFIM_DE")
	require.Contains(t, portugueseDemoExpectedRuleIDs_ptBR, "AFIM_DE")
}

// Analyze demo text (runTests-adjacent surface) — real PT sentence split.
func TestPortuguese_AnalyzeDemoText(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-PT")
	sents := lt.Analyze(portugueseDemoText)
	require.NotEmpty(t, sents)
	require.GreaterOrEqual(t, len(sents), 2)
}

// pt-BR core demo surface: same core hits, BR speller id + Brazilian extras wired.
func TestPortuguese_Language_BrazilianCoreDemo(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-BR")
	RegisterCorePortugueseRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_BR")
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "PT_BR_SIMPLE_REPLACE")
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "PT_POSAO_DASH_RULE")
	// unique set: second instances share getId with base packs
	uniq := func(ids []string) []string {
		seen := map[string]struct{}{}
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
		return out
	}
	require.ElementsMatch(t, uniq(language.BrazilianPortugueseRelevantRuleIDs()), uniq(lt.GetAllRegisteredRuleIDs()))

	matches := lt.Check(portugueseDemoText)
	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		ids = append(ids, m.RuleID)
	}
	require.Contains(t, ids, "PORTUGUESE_WORD_REPEAT_RULE")
	spellerHits := 0
	for _, id := range ids {
		if id == "MORFOLOGIK_RULE_PT_BR" {
			spellerHits++
		}
	}
	require.GreaterOrEqual(t, spellerHits, 1,
		"demo typos must hit MORFOLOGIK_RULE_PT_BR when pt-BR dict loads")
}

// pt-AO core demo surface: PreReform extras + AO speller id; no PT/BR packs.
func TestPortuguese_Language_AngolaCoreDemo(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-AO")
	RegisterCorePortugueseRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_PT_AO")
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "PT_COMPOUNDS_PRE_REFORM")
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "PT_PREAO_DASH_RULE")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "PT_POSAO_DASH_RULE")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "PT_BR_SIMPLE_REPLACE")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "PT_PT_SIMPLE_REPLACE")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "PT_AGREEMENT_REPLACE")
	uniq := func(ids []string) []string {
		seen := map[string]struct{}{}
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
		return out
	}
	require.ElementsMatch(t, uniq(language.AngolaPortugueseRelevantRuleIDs()), uniq(lt.GetAllRegisteredRuleIDs()))

	matches := lt.Check(portugueseDemoText)
	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		ids = append(ids, m.RuleID)
	}
	require.Contains(t, ids, "PORTUGUESE_WORD_REPEAT_RULE")
	spellerHits := 0
	for _, id := range ids {
		if id == "MORFOLOGIK_RULE_PT_AO" {
			spellerHits++
		}
	}
	require.GreaterOrEqual(t, spellerHits, 1,
		"demo typos must hit MORFOLOGIK_RULE_PT_AO when pt-AO dict loads")
}
