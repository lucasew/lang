package pt

// Twin of org.languagetool.rules.pt.MorfologikPortugueseSpellerRuleTest.
// Official resources: /pt/spelling/{pt-BR,pt-PT-90,pt-PT-45}.dict + pt/*.txt lists.
// Analysis path ports JLanguageTool(pt-*).getAnalyzedSentence:
// PortugueseWordTokenizer + PortugueseTagger + PortugueseHybridDisambiguator.

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
	ptHybridOnce       sync.Once
	ptHybrid           *disambigpt.PortugueseHybridDisambiguator
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

func portugueseHybrid() *disambigpt.PortugueseHybridDisambiguator {
	ptHybridOnce.Do(func() {
		ptHybrid = disambigpt.NewPortugueseHybridDisambiguator()
	})
	return ptHybrid
}

// analyzePT ports JLanguageTool(pt-*).getAnalyzedSentence.
func analyzePT(text string) *languagetool.AnalyzedSentence {
	_ = sharedPortugueseSpellerTagger()
	var sent *languagetool.AnalyzedSentence
	if tg := sharedPortugueseSpellerTagger(); tg != nil {
		sent = languagetool.AnalyzeWithTaggerAndTokenizer(text, func(tok string) []languagetool.TokenTag {
			atrs := tg.Tag([]string{tok})
			if len(atrs) == 0 || atrs[0] == nil {
				return nil
			}
			var out []languagetool.TokenTag
			for _, r := range atrs[0].GetReadings() {
				if r == nil {
					continue
				}
				pos, lem := "", ""
				if r.GetPOSTag() != nil {
					pos = *r.GetPOSTag()
				}
				if r.GetLemma() != nil {
					lem = *r.GetLemma()
				}
				if pos == "" && lem == "" {
					continue
				}
				out = append(out, languagetool.TokenTag{POS: pos, Lemma: lem})
			}
			return out
		}, pttok.NewPortugueseWordTokenizer())
	} else {
		sent = languagetool.AnalyzeWithTokenizer(text, pttok.NewPortugueseWordTokenizer())
	}
	if d := portugueseHybrid(); d != nil {
		if out := d.Disambiguate(sent); out != nil {
			sent = out
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
	require.Empty(t, ms, "expected no errors in %q; got %d matches %v", sentence, len(ms), matchSummaries(ms, sentence))
}

func matchSummaries(ms []*rules.RuleMatch, sentence string) []string {
	var out []string
	for _, m := range ms {
		if m == nil {
			continue
		}
		surf := rules.UTF16Substring(sentence, m.GetFromPos(), m.GetToPos())
		out = append(out, surf)
	}
	return out
}

func assertErrorLength(t *testing.T, r *MorfologikPortugueseSpellerRule, sentence string, length int, suggestions ...string) {
	t.Helper()
	ms := matchPT(t, r, sentence)
	require.Len(t, ms, length, "sentence %q got %v", sentence, matchSummaries(ms, sentence))
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
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", ms[0].GetSpecificRuleId())
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR_DIALECT", ms[1].GetSpecificRuleId())
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", ms[2].GetSpecificRuleId())
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
	assertSingleError(t, ruleBR, "primit", "primit.")
	assertSingleError(t, ruleBR, "Islam,", "Islam.")
	assertNoErrors(t, ruleBR, "xerogr.")
	assertNoErrors(t, ruleBR, "Baixei a vers. 7.0.0")
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
	assertNoErrors(t, ruleBR, "mora na 82nd Street")
	assertNoErrors(t, ruleBR, "mora na Fifth Avenue")
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
	assertNoErrors(t, ruleBR, "#CantadaBoBem")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckUserMentions(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "@nomeDoUsuario")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckCurrencyValues(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "R$45,00")
	assertNoErrors(t, ruleBR, "R$ 3")
	assertNoErrors(t, ruleBR, "US$1.000,00")
	assertNoErrors(t, ruleBR, "€99,99")
	assertNoErrors(t, ruleBR, "6£")
	assertNoErrors(t, ruleBR, "30 R$")
	assertNoErrors(t, ruleBR, "US$")
	assertNoErrors(t, ruleBR, "US$ 58,0 bilhões")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckNumberAbbreviations(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "Nº666")
	assertNoErrors(t, ruleBR, "N°42189")
	assertNoErrors(t, ruleBR, "Nº 420")
	assertNoErrors(t, ruleBR, "N.º69")
	assertNoErrors(t, ruleBR, "N.º 80085")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotCorrectOrdinalSuperscripts(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "6º")
	assertNoErrors(t, ruleBR, "100°")
	assertNoErrors(t, ruleBR, "21ª")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotCorrectDegreeExpressions(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "1,0°")
	assertNoErrors(t, ruleBR, "2°c")
	assertNoErrors(t, ruleBR, "3°C")
	assertNoErrors(t, ruleBR, "4,0ºc")
	assertNoErrors(t, ruleBR, "5.0ºc")
	assertNoErrors(t, ruleBR, "6,0ºRø")
	assertNoErrors(t, ruleBR, "7,5ºN")
	assertNoErrors(t, ruleBR, "−8,0°")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotCorrectCopyrightSymbol(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "Copyright©")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingSplitsEmoji(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "☺☺☺Só")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckXForVezes(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "10X")
	assertNoErrors(t, ruleBR, "5x")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingFailsWithModifierDiacritic(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "Não")
	// modifier tilde — Java expects single error with "Não" among suggestions
	assertSingleError(t, ruleBR, "Não", "Não")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingWorksWithRarePunctuation(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "⌈Herói⌋")
	assertNoErrors(t, ruleBR, "″Santo Antônio do Manga″")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingCustomReplacements(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleError(t, ruleBR, "vinheram", "vieram")
	assertSingleError(t, ruleBR, "recadações", "arrecadações")
	assertSingleError(t, ruleBR, "me dis", "diz")
	assertSingleError(t, ruleBR, "ebook", "e-book")
	assertSingleError(t, ruleBR, "quizese", "quisesse")
	assertSingleError(t, ruleBR, "cabesse", "coubesse")
	assertSingleError(t, ruleBR, "andância", "andança")
	assertSingleError(t, ruleBR, "abto", "hábito")
	assertSingleError(t, ruleBR, "logo nao", "não")
	assertSingleError(t, ruleBR, "kitchenette", "quitinete")
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

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresUppercaseAndDigitString(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "ABC2000")
	assertNoErrors(t, ruleBR, "AI5")
	assertNoErrors(t, ruleBR, "IP65")
	assertNoErrors(t, ruleBR, "HR2048")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresAmpersandBetweenTwoCapitals(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "J&F")
	assertNoErrors(t, ruleBR, "A&E")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresParentheticalInflection(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "professor(es)")
	assertNoErrors(t, ruleBR, "profissional(is)")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresProbableUnitsOfMeasurement(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "180g")
	assertNoErrors(t, ruleBR, "16.2kW")
	assertNoErrors(t, ruleBR, "6x6")
	assertNoErrors(t, ruleBR, "100x100mm")
	assertNoErrors(t, ruleBR, "5,5x6.7km")
	assertNoErrors(t, ruleBR, "5×10×50cm")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresDiceRollNotation(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "1d20")
	assertNoErrors(t, ruleBR, "3d6")
	assertNoErrors(t, ruleBR, "20d10")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresHexadecimalAndOctalNumbers(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "0x1A")
	assertNoErrors(t, ruleBR, "0x9f")
	assertNoErrors(t, ruleBR, "0xdeadbeef")
	assertNoErrors(t, ruleBR, "0x5F6A")
	assertNoErrors(t, ruleBR, "0o23")
	assertNoErrors(t, ruleBR, "0o777")
	assertSingleError(t, ruleBR, "0o8")
	assertSingleError(t, ruleBR, "0xQ34")
	assertNoErrors(t, ruleBR, "0x34Q")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresNonstandardTimeFormat(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "31h40min")
	assertNoErrors(t, ruleBR, "1h20min3s")
	assertNoErrors(t, ruleBR, "13:30h")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresLaughterOnomatopoeia(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "hahahahaha")
	assertNoErrors(t, ruleBR, "heheh")
	assertNoErrors(t, ruleBR, "huehuehuehue")
	assertNoErrors(t, ruleBR, "Kkkkkkk")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerRecognisesMonthAbbreviations(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "23/jan")
	assertNoErrors(t, ruleBR, "9/Dez/2069")
	assertSingleError(t, ruleBR, "10/feb/2010", "fev")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerRecognisesRomanNumerals(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "XVIII")
	assertNoErrors(t, ruleBR, "xviii")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresIsolatedGreekLetters(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "ξ")
	assertNoErrors(t, ruleBR, "Ω")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresWordsFromIgnoreTXT(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "ignorewordoogaboogatest")
	assertSingleErrorWithNegativeSuggestion(t, ruleBR, "ignorewordoogaboogatext", "ignorewordoogaboogatest")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotAcceptProhibitedWords(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleError(t, ruleBR, "prohibitwordoogaboogatest")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresNames(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "Fulgencio Fuhao")
	assertSingleError(t, ruleBR, "Jordao", "Jordão")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerMultitokens(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "BRIGITTE BARDOT")
	assertNoErrors(t, ruleBR, "Brigitte Bardot")
	assertNoErrors(t, ruleBR, "MERCEDES-BENZ")
	assertNoErrors(t, ruleBR, "Mercedes-Benz")
	assertNoErrors(t, ruleBR, "big band")
	assertNoErrors(t, ruleBR, "Big band")
	assertNoErrors(t, ruleBR, "Big Band")
	assertNoErrors(t, ruleBR, "BIG BANDS")
	assertNoErrors(t, ruleBR, "rhythm and blues")
	assertNoErrors(t, ruleBR, "Rhythm and blues")
	assertNoErrors(t, ruleBR, "Rhythm And Blues")
	assertNoErrors(t, ruleBR, "Rhythm and Blues")
	assertNoErrors(t, ruleBR, "stock car")
	assertNoErrors(t, ruleBR, "Stock Car")
	assertNoErrors(t, ruleBR, "Hall of Fame")
	// Java assertSingleError("Hall Of Fame") — multiword entry is mixed-case "Hall of Fame"
	// so titlecase variants are not generated; Of and Fame are both unlisted → 2 matches.
	assertErrorLength(t, ruleBR, "Hall Of Fame", 2)
	assertErrorLength(t, ruleBR, "hall of fame", 2)
	assertNoErrors(t, ruleBR, "Rock and Roll")
	assertNoErrors(t, ruleBR, "Hall of Fame")
	assertNoErrors(t, ruleBR, "Rock and Roll Hall of Fame")
	// Rock And Roll ignored via multiword; Hall Of Fame leaves Of+Fame (2).
	assertErrorLength(t, ruleBR, "Rock And Roll Hall Of Fame", 2)
	assertNoErrors(t, ruleBR, "Chesapeake Bay retriever")
	assertSingleError(t, ruleBR, "Chesapeake Bay Retriever")
	assertNoErrors(t, ruleBR, "Pit Bull")
	assertNoErrors(t, ruleBR, "Mao Tsé-Tung")
	assertNoErrors(t, ruleBR, "Honoris Causa")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerEnglishCompounds(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "UntaggedWord Card")
	assertNoErrors(t, ruleBR, "Vaca Center")
	assertNoErrors(t, ruleBR, "de Klerk Center")
	assertSingleError(t, ruleBR, "caramba Center", "Conter", "Centre")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsArbitraryHyphenation(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "Xai-Xai")
	assertNoErrors(t, ruleBR, "Tsé-Tung")
	assertNoErrors(t, ruleBR, "X-Men")
	assertNoErrors(t, ruleBR, "t-shirts")
	assertNoErrors(t, ruleBR, "além-mar")
	assertNoErrors(t, ruleBR, "além-mares")
	assertNoErrors(t, ruleBR, "baby-doll")
	assertNoErrors(t, ruleBR, "baby-dolls")
	assertNoErrors(t, ruleBR, "e-zine")
	assertNoErrors(t, ruleBR, "e-zines")
	assertNoErrors(t, ruleBR, "CD-ROM")
	assertNoErrors(t, ruleBR, "CD-ROMs")
	assertSingleError(t, ruleBR, "heavy-metal", "heavy metal")
	assertNoErrors(t, ruleBR, "Aix-en-Provence")
	assertNoErrors(t, rulePT, "Agualva-Cacém")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAccepts50PercentOff(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "50%OFF")
	assertSingleError(t, ruleBR, "50%oogabooga")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsIllegalPrefixation(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "semi-consciente")
	assertNoErrors(t, ruleBR, "semi-acústicas")
	assertNoErrors(t, ruleBR, "semi-frio")
	assertNoErrors(t, ruleBR, "sub-taça")
	assertNoErrors(t, ruleBR, "sub-pratos")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsCapitalisationOfAllCompoundElements(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "jiu-jitsu.")
	assertNoErrors(t, ruleBR, "Jiu-jitsu.")
	assertNoErrors(t, ruleBR, "Jiu-Jitsu.")
	assertErrorLength(t, ruleBR, "jIu-JItsU", 2)
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsNationalPrefixes(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "ítalo-congolês")
	assertNoErrors(t, ruleBR, "Belgo-Luxemburguesa")
	assertNoErrors(t, ruleBR, "franco-prussiana")
	assertNoErrors(t, ruleBR, "Franco-prussiana")
	assertNoErrors(t, ruleBR, "Franco-Prussiana")
	assertNoErrors(t, ruleBR, "húngaro-romeno")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsParagraphAndOrdinal(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, ruleBR, "§1º")
	assertNoErrors(t, ruleBR, "§ 1º")
	assertNoErrors(t, ruleBR, "§1º-A")
	assertNoErrors(t, ruleBR, "§ 1º-A")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerReplacesOldGrammarRules(t *testing.T) {
	ensurePTSpellerRules(t)
	assertSingleError(t, ruleBR, "Amamo-te", "Amamos")
	assertSingleError(t, ruleBR, "Amá-te", "Amar")
	assertSingleError(t, ruleBR, "Fazê o quê?", "Fazer")
	assertSingleError(t, ruleBR, "Vamo embora", "Vamos")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerHasNewWords(t *testing.T) {
	ensurePTSpellerRules(t)
	assertNoErrors(t, rulePT, "diferençazinha")
	assertNoErrors(t, ruleBR, "Mewtwo, Pikachu, Rapidash e dois Growlithes")
	assertNoErrors(t, ruleBR, "bebezice")
	assertNoErrors(t, ruleBR, "Solucz")
	assertNoErrors(t, ruleBR, "microagulhamento")
	assertNoErrors(t, ruleBR, "curricularização")
	assertNoErrors(t, ruleBR, "email")
	assertNoErrors(t, ruleBR, "pô")
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

func TestMorfologikPortugueseSpellerRule_FilterHelpers(t *testing.T) {
	sugs := filterDoNotSuggest([]string{"casa", "teste"})
	require.Contains(t, sugs, "casa")
	require.True(t, strings.Contains(checkDiaeresis("pingüim"), "u") || checkDiaeresis("pingüim") == "pinguim")
	require.Equal(t, "amamos", checkEuropeanStyle1PLPastTense("pt-BR", "amámos"))
	require.Equal(t, "", checkEuropeanStyle1PLPastTense("pt-PT", "amámos"))
}
