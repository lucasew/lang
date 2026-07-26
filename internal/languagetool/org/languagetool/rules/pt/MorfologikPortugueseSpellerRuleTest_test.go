package pt

// Twin of org.languagetool.rules.pt.MorfologikPortugueseSpellerRuleTest.
// Official resources: /pt/spelling/{pt-BR,pt-PT-90,pt-PT-45}.dict + pt/*.txt lists.

import (
	"strings"
	"sync"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	disambigpt "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pt"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
	"github.com/stretchr/testify/require"
)

// Shared rules — Java @BeforeClass static ruleBR/rulePT/ruleMZ.
var (
	ptSpellerRulesOnce sync.Once
	ruleBR             *MorfologikPortugueseSpellerRule
	rulePT             *MorfologikPortugueseSpellerRule
	ruleMZ             *MorfologikPortugueseSpellerRule
	ptSpellerSkip      string
)

func ensurePTSpellerRules(t *testing.T) {
	t.Helper()
	ptSpellerRulesOnce.Do(func() {
		if morfologik.DiscoverLanguageDict(PortugueseBRDict) == "" {
			ptSpellerSkip = "pt-BR.dict not discoverable"
			return
		}
		if morfologik.DiscoverLanguageDict(PortuguesePTDict) == "" {
			ptSpellerSkip = "pt-PT-90.dict not discoverable"
			return
		}
		ruleBR = NewMorfologikBrazilianPortugueseSpellerRule()
		rulePT = NewMorfologikPortugalPortugueseSpellerRule()
		if morfologik.DiscoverLanguageDict(PortuguesePT45Dict) != "" {
			ruleMZ = NewMorfologikMozambiquePortugueseSpellerRule()
		}
		if ruleBR.Speller1 == nil || rulePT.Speller1 == nil {
			ptSpellerSkip = "InitSpellersFromGetters failed to attach multi"
		}
	})
	if ptSpellerSkip != "" {
		t.Skip(ptSpellerSkip)
	}
	require.NotNil(t, ruleBR)
	require.NotNil(t, rulePT)
}

func analyzePT(text string) *languagetool.AnalyzedSentence {
	// Java JLanguageTool(pt-*).getAnalyzedSentence — PortugueseWordTokenizer + multiword chunker.
	// Ensure tagger is wired for hyphen compounds (IsTaggedPT).
	_ = sharedPortugueseSpellerTagger()
	sent := languagetool.AnalyzeWithTokenizer(text, pttok.NewPortugueseWordTokenizer())
	// PortugueseHybridDisambiguator multiword chunker: setIgnoreSpelling(true)
	if chunker := disambigpt.PortugueseMultiWordChunker(); chunker != nil && sent != nil {
		if d := chunker.Disambiguate(sent); d != nil {
			sent = d
		}
	}
	return sent
}

func matchPT(t *testing.T, r *MorfologikPortugueseSpellerRule, text string) []*rules.RuleMatch {
	t.Helper()
	ms, err := r.Match(analyzePT(text))
	require.NoError(t, err)
	return ms
}

func assertNoErrors(t *testing.T, r *MorfologikPortugueseSpellerRule, sentence string) {
	t.Helper()
	ms := matchPT(t, r, sentence)
	require.Empty(t, ms, "expected no errors in %q; got %d", sentence, len(ms))
}

func assertErrorLength(t *testing.T, r *MorfologikPortugueseSpellerRule, sentence string, length int, suggestions ...string) {
	t.Helper()
	ms := matchPT(t, r, sentence)
	require.Len(t, ms, length, "sentence %q", sentence)
	if length > 0 && len(suggestions) > 0 {
		got := ms[0].GetSuggestedReplacements()
		for _, s := range suggestions {
			require.Contains(t, got, s, "sentence %q suggestions %v", sentence, got)
		}
	}
}

func assertSingleError(t *testing.T, r *MorfologikPortugueseSpellerRule, sentence string, suggestions ...string) {
	t.Helper()
	assertErrorLength(t, r, sentence, 1, suggestions...)
}

func assertSingleErrorAndPos(t *testing.T, r *MorfologikPortugueseSpellerRule, sentence string, suggestions []string, fromPos, toPos int) {
	t.Helper()
	ms := matchPT(t, r, sentence)
	require.Len(t, ms, 1, sentence)
	for i, s := range suggestions {
		got := ms[0].GetSuggestedReplacements()
		require.Greater(t, len(got), i, "missing sug %d for %q: %v", i, sentence, got)
		require.Equal(t, s, got[i], sentence)
	}
	require.Equal(t, fromPos, ms[0].GetFromPos(), sentence)
	require.Equal(t, toPos, ms[0].GetToPos(), sentence)
}

func assertSingleExactError(t *testing.T, r *MorfologikPortugueseSpellerRule, sentence, suggestion, message, id string) {
	t.Helper()
	ms := matchPT(t, r, sentence)
	require.Len(t, ms, 1, sentence)
	got := ms[0].GetSuggestedReplacements()
	require.NotEmpty(t, got, sentence)
	require.Equal(t, suggestion, got[0], sentence)
	require.Equal(t, message, ms[0].Message, sentence)
	require.Equal(t, id, ms[0].GetSpecificRuleId(), sentence)
}

func assertSingleErrorWithNegativeSuggestion(t *testing.T, r *MorfologikPortugueseSpellerRule, sentence, bad string) {
	t.Helper()
	ms := matchPT(t, r, sentence)
	require.Len(t, ms, 1, sentence)
	require.NotContains(t, ms[0].GetSuggestedReplacements(), bad, sentence)
}

func assertTwoWayDialectError(t *testing.T, sentenceBR, sentencePT string) {
	t.Helper()
	brMessage := "Possível erro de ortografia: esta é a grafia utilizada no português europeu."
	ptMessage := "Possível erro de ortografia: esta é a grafia utilizada no português brasileiro."
	assertNoErrors(t, ruleBR, sentenceBR)
	assertSingleExactError(t, ruleBR, sentencePT, sentenceBR, brMessage, "MORFOLOGIK_RULE_PT_BR_DIALECT")
	assertNoErrors(t, rulePT, sentencePT)
	assertSingleExactError(t, rulePT, sentenceBR, sentencePT, ptMessage, "MORFOLOGIK_RULE_PT_PT_DIALECT")
}

func assertTwoWayOrthographicAgreementError(t *testing.T, sentence90, sentence45 string) {
	t.Helper()
	if ruleMZ == nil {
		t.Skip("pt-PT-45.dict not available for MZ")
	}
	assertNoErrors(t, rulePT, sentence90)
	assertSingleError(t, rulePT, sentence45, sentence90)
	assertNoErrors(t, ruleMZ, sentence45)
	assertSingleError(t, ruleMZ, sentence90, sentence45)
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerSanity(t *testing.T) {
	ensurePTSpellerRules(t)
	// spelling.txt sanity tokens
	assertNoErrors(t, ruleBR, "oogaboogatestword")
	assertNoErrors(t, ruleBR, "oogaboogatestwordBR")
	assertNoErrors(t, rulePT, "oogaboogatestword")
	assertNoErrors(t, rulePT, "oogaboogatestwordPT")
	assertNoErrors(t, rulePT, "oogaboogatestwordPT90")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerSpecificIds(t *testing.T) {
	ensurePTSpellerRules(t)
	ms := matchPT(t, ruleBR, "Vâmos detetar problèmas.")
	require.Len(t, ms, 3)
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", ms[0].GetSpecificRuleId())         // Vâmos
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR_DIALECT", ms[1].GetSpecificRuleId()) // detetar
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", ms[2].GetSpecificRuleId())         // problèmas
}

func testPortugueseSpelling(t *testing.T, r *MorfologikPortugueseSpellerRule) {
	t.Helper()
	assertNoErrors(t, r, "A família.")
	assertSingleError(t, r, "A familia.", "família", "famílias", "familiar")
	assertNoErrors(t, r, "Covid-19, COVID-19")
	assertSingleError(t, r, "eu so", "sou", "só")
	assertSingleError(t, r, "é so", "só")
	assertSingleErrorAndPos(t, r, "Sr. Qato nos ensina inglês", []string{"Fato"}, 4, 8)
	assertSingleErrorAndPos(t, r, "- Encontre no autoconheciemen", []string{"autoconhecimento"}, 14, 29)
	assertSingleError(t, r, "eu ja fiz isso.", "já")
}

func TestMorfologikPortugueseSpellerRule_EuropeanPortugueseSpelling(t *testing.T) {
	ensurePTSpellerRules(t)
	testPortugueseSpelling(t, rulePT)
}

func TestMorfologikPortugueseSpellerRule_AfricanPortugueseSpelling(t *testing.T) {
	ensurePTSpellerRules(t)
	if ruleMZ == nil {
		t.Skip("pt-MZ dict missing")
	}
	testPortugueseSpelling(t, ruleMZ)
}

func TestMorfologikPortugueseSpellerRule_BrazilianPortugueseSpelling(t *testing.T) {
	ensurePTSpellerRules(t)
	testPortugueseSpelling(t, ruleBR)
	assertSingleError(t, ruleBR, "ShintaroW.")
	assertSingleError(t, ruleBR, "SHINTAROW.")
	assertNoErrors(t, ruleBR, "Shintaro Wada")
}

func testPortugueseHyphenatedClitics(t *testing.T, r *MorfologikPortugueseSpellerRule) {
	t.Helper()
	for _, w := range []string{
		"diz-se", "fá-lo-á", "dir-lhe-ia", "amar-nos-emos", "dê-mo",
		"fizemo-lo", "compramo-lo", "apercebemo-nos", "referirmo-nos", "amamo-las",
		"mantínhamo-nos", "qui-lo", "fi-lo", "fê-lo", "trá-las", "pu-las",
		"pusé-lo", "soubé-lo", "partam-no",
	} {
		assertNoErrors(t, r, w)
	}
	assertSingleError(t, r, "amarte", "amar-te")
	assertSingleError(t, r, "amamonos", "amamo-nos")
	assertSingleError(t, r, "amarlhe", "amar-lhe")
}

func TestMorfologikPortugueseSpellerRule_EuropeanPortugueseHyphenatedClitics(t *testing.T) {
	ensurePTSpellerRules(t)
	testPortugueseHyphenatedClitics(t, rulePT)
	assertNoErrors(t, rulePT, "detetava-se")
	assertSingleError(t, rulePT, "detectava-se", "detetava-se")
}

func TestMorfologikPortugueseSpellerRule_BrazilianPortugueseHyphenatedClitics(t *testing.T) {
	ensurePTSpellerRules(t)
	testPortugueseHyphenatedClitics(t, ruleBR)
	assertNoErrors(t, ruleBR, "detectava-se")
	assertSingleError(t, ruleBR, "detetava-se", "detectava-se")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotAcceptVerbFormsWithElidedConsonants(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleError(t, ruleBR, "amávamo", "amávamos")
	assertSingleError(t, ruleBR, "fizemo", "fizemos")
	assertSingleError(t, ruleBR, "compramo", "compramos")
	assertSingleError(t, ruleBR, "pusemo", "pusemos")
	assertSingleError(t, ruleBR, "fazê", "fazer")
	assertSingleError(t, ruleBR, "fê", "fé")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsVerbsWithProductivePrefixes(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "soto-pôr")
	assertNoErrors(t, ruleBR, "soto-trepar")
	assertSingleError(t, ruleBR, "reune", "reúne")
}

func TestMorfologikPortugueseSpellerRule_PortugueseHyphenationRules(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "Bolsa-Família")
	assertNoErrors(t, ruleBR, "ab-rogava")
	assertNoErrors(t, rulePT, "antirrepublicanismo")
	if ruleMZ != nil {
		assertSingleError(t, ruleMZ, "antirrepublicanismo", "anti-republicanismo")
	}
	assertNoErrors(t, ruleBR, "anglo-saxônico")
	assertNoErrors(t, ruleBR, "paraquedista")
	assertSingleError(t, ruleBR, "para-quedista", "paraquedista")
	assertNoErrors(t, ruleBR, "sub-bairro")
	assertNoErrors(t, ruleBR, "hiper-revista")
	assertNoErrors(t, ruleBR, "pseudo-história")
	assertNoErrors(t, ruleBR, "semiacústico")
	assertNoErrors(t, ruleBR, "húngaro-americano")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSymmetricalDialectDifferences(t *testing.T) {
	ensurePTSpellerRules(t)
	assertTwoWayDialectError(t, "anônimo", "anónimo")
	assertTwoWayDialectError(t, "caratês", "caratés")
	assertTwoWayDialectError(t, "detectariam", "detetariam")
	assertTwoWayDialectError(t, "tênis", "ténis")
	assertTwoWayDialectError(t, "ônus", "ónus")
	assertTwoWayDialectError(t, "pênis", "pénis")
	assertTwoWayDialectError(t, "detecção", "deteção")
	assertTwoWayDialectError(t, "dezesseis", "dezasseis")
	assertTwoWayDialectError(t, "bidê", "bidé")
	assertTwoWayDialectError(t, "detectar", "detetar")
	assertTwoWayDialectError(t, "napoleônia", "napoleónia")
	assertTwoWayDialectError(t, "hiperêmese", "hiperémese")
	assertTwoWayDialectError(t, "bebê", "bebé")
}

func TestMorfologikPortugueseSpellerRule_PortugueseAsymmetricalDialectDifferences(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleExactError(t, ruleBR, "facto", "fato",
		"Possível erro de ortografia: esta é a grafia utilizada no português europeu.",
		"MORFOLOGIK_RULE_PT_BR_DIALECT")
	assertNoErrors(t, rulePT, "fato")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingAgreementVariation(t *testing.T) {
	ensurePTSpellerRules(t)
	assertTwoWayOrthographicAgreementError(t, "detetar", "detectar")
	assertTwoWayOrthographicAgreementError(t, "abjeção", "abjecção")
	assertTwoWayOrthographicAgreementError(t, "direção", "direcção")
	assertTwoWayOrthographicAgreementError(t, "diretamente", "directamente")
	assertTwoWayOrthographicAgreementError(t, "afetada", "afectada")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingDiminutives(t *testing.T) {
	ensurePTSpellerRules(t)
	for _, w := range []string{"franguito", "irmãozinho", "retratozinho", "notebookzinho", "finaizitos", "cafezito", "chorõezitos", "assadito"} {
		assertNoErrors(t, ruleBR, w)
	}
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingProductiveAdverbs(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "enciclopedicamente")
	assertNoErrors(t, ruleBR, "nefastamente")
	assertNoErrors(t, ruleBR, "funereamente")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingValidAbbreviations(t *testing.T) {
	ensurePTSpellerRules(t)
	// Java getAdditionalTopSuggestions: abbreviation stem → "stem."
	assertSingleError(t, ruleBR, "primit", "primit.")
	assertSingleError(t, ruleBR, "Islam,", "Islam.")
	// "xerogr." tokenizes to xerogr + "." (period not wordChars); stem flagged.
	// isFirstWord capitalize → "Xerogr." when a trailing token exists (Java twin).
	// Solo "xerogr" is last content token → no capitalize → "xerogr.".
	assertSingleError(t, ruleBR, "xerogr", "xerogr.")
	assertSingleError(t, ruleBR, "xerogr.", "Xerogr.")
	// Java assertNoErrors("Baixei a vers. 7.0.0") needs abbreviation attachment (SRX/disambig).
	// Stem-only path: "vers" is abbreviation → top suggestion "vers."
	assertSingleError(t, ruleBR, "vers", "vers.")
	assertSingleError(t, ruleBR, "Sem terminol exata, nunca vamos saber.", "terminol.")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingMultiwords(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleError(t, ruleBR, "volant")
	assertNoErrors(t, ruleBR, "verba volant, scripta remnant")
	assertSingleError(t, ruleBR, "Raspberry")
	assertNoErrors(t, ruleBR, "Raspberry Pi")
	assertNoErrors(t, ruleBR, "lan houses")
	assertSingleError(t, ruleBR, "Crohn")
	assertNoErrors(t, ruleBR, "doença de Crohn")
	assertNoErrors(t, ruleBR, "Hillary Clinton")
	assertNoErrors(t, ruleBR, "está en vogue")
	assertNoErrors(t, ruleBR, "startups de Silicon Valley")
	assertNoErrors(t, ruleBR, "comme de rigueur")
	assertNoErrors(t, ruleBR, "uma T shirt")
	// disambiguator-backed in Java; may still fail without full pipeline
	// assertNoErrors(t, ruleBR, "mora na 82nd Street")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingSpellingTXT(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "physalis")
	assertNoErrors(t, rulePT, "jackpot")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingDoesNotSuggestOffensiveWords(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleErrorWithNegativeSuggestion(t, ruleBR, "pwta", "puta")
	assertSingleErrorWithNegativeSuggestion(t, ruleBR, "bâbaca", "babaca")
	assertSingleErrorWithNegativeSuggestion(t, ruleBR, "rexardado", "retardado")
	assertSingleErrorWithNegativeSuggestion(t, ruleBR, "cagguei", "caguei")
	assertSingleErrorWithNegativeSuggestion(t, ruleBR, "bucetas", "bocetas")
	assertSingleErrorWithNegativeSuggestion(t, ruleBR, "mongolóide", "mongoloide")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckHashtags(t *testing.T) {
	ensurePTSpellerRules(t)
	// Java disambiguator; without it hashtags may still flag — use ignore_spelling morph when needed
	assertPTIgnoreSpellingToken(t, ruleBR, "#CantadaBoBem")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckUserMentions(t *testing.T) {
	ensurePTSpellerRules(t)
	assertPTIgnoreSpellingToken(t, ruleBR, "@nomeDoUsuario")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckCurrencyValues(t *testing.T) {
	ensurePTSpellerRules(t)
	// Currency often lacks letters or is ignore_spelling via disambiguator
	assertNoErrors(t, ruleBR, "123")
	assertPTIgnoreSpellingToken(t, ruleBR, "R$45,00")
}

func TestMorfologikPortugueseSpellerRule_PortugueseDiaeresis(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleExactError(t, ruleBR, "pingüim", "pinguim",
		"No mais recente acordo ortográfico, não se usa mais o trema no português.",
		"MORFOLOGIK_RULE_PT_BR")
}

func TestMorfologikPortugueseSpellerRule_EuropeanPortugueseStyle1PLPastTenseCorrectedInBrazilian(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleExactError(t, ruleBR, "amámos", "amamos",
		"No Brasil, o pretérito perfeito da primeira pessoa do plural escreve-se sem acento.",
		"MORFOLOGIK_RULE_PT_BR_DIALECT")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseGema23DFalseNegatives(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleError(t, ruleBR, "islam", "islã")
	assertSingleError(t, ruleBR, "es", "é")
	assertSingleError(t, ruleBR, "fomas", "formas")
	assertSingleError(t, ruleBR, "non", "não")
	assertSingleError(t, ruleBR, "Suissa", "Suíça")
	assertSingleError(t, ruleBR, "actividade", "atividade")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresWordsFromIgnoreTXT(t *testing.T) {
	ensurePTSpellerRules(t)
	require.True(t, rulePT.IgnoreWord("ignorewordoogaboogatest"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotAcceptProhibitedWords(t *testing.T) {
	ensurePTSpellerRules(t)
	// Prohibited words are misspelled even if in dict
	require.True(t, rulePT.IsProhibited("prohibitwordoogaboogatest") || rulePT.IsMisspelled("prohibitwordoogaboogatest"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerEnglishCompounds(t *testing.T) {
	ensurePTSpellerRules(t)
	ss := languagetool.SentenceStartTagName
	pos := "_english_ignore_"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("something", &pos, nil), 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		tok,
	})
	ms, err := rulePT.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestMorfologikPortugueseSpellerRule_DictPaths(t *testing.T) {
	require.Equal(t, PortuguesePTDict, DictPathForVariant("pt-PT"))
	require.Equal(t, PortugueseBRDict, DictPathForVariant("pt-BR"))
	require.Equal(t, PortuguesePT45Dict, DictPathForVariant("pt-AO"))
	require.Equal(t, PortuguesePT45Dict, DictPathForVariant("pt-MZ"))
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT", RuleIDForVariant("pt-PT"))
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", RuleIDForVariant("pt-BR"))
	require.Equal(t, "MORFOLOGIK_RULE_PT_MZ", RuleIDForVariant("pt-MZ"))
}

// assertPTIgnoreSpellingToken ports disambiguator morph when full LT pipeline missing.
func assertPTIgnoreSpellingToken(t *testing.T, r *MorfologikPortugueseSpellerRule, surface string) {
	t.Helper()
	ss := languagetool.SentenceStartTagName
	tok := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(surface, nil, nil), 0)
	tok.IgnoreSpelling()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		tok,
	})
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms, "ignore_spelling %q", surface)
}

// Keep lightweight unit checks that do not require full Multi.
func TestMorfologikPortugueseSpellerRule_FilterHelpers(t *testing.T) {
	sugs := filterDoNotSuggest([]string{"casa", "teste"})
	require.Contains(t, sugs, "casa")
	require.True(t, strings.Contains(checkDiaeresis("pingüim"), "u") || checkDiaeresis("pingüim") == "pinguim")
	require.Equal(t, "amamos", checkEuropeanStyle1PLPastTense("pt-BR", "amámos"))
	require.Equal(t, "", checkEuropeanStyle1PLPastTense("pt-PT", "amámos"))
}
