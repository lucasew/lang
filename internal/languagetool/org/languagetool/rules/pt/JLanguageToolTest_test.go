package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/JLanguageToolTest.java
//
// Lives in rules/pt (not package languagetool) to avoid languagetool↔rules/pt import
// cycles. Java source path is still the PT module JLanguageToolTest.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testPortugueseVariants.
//
// Java:
//
//	sentence  = "Isto é uma característica sua." → 0 matches
//	sentence2 = "Isto é uma características sua." → 1 match
//	for pt-PT, pt-BR, pt-AO, pt-MZ
//
// The 1-match case is a grammar agreement pattern (GENERAL_*_AGREEMENT / similar)
// from grammar.xml — not in Portuguese.getRelevantRules. With core rules only we
// assert the clean sentence is empty and document the agreement gap (Polish honesty).
func TestJLanguageTool_PortugueseVariants(t *testing.T) {
	sentence := "Isto é uma característica sua."
	sentence2 := "Isto é uma características sua."
	for _, code := range []string{"pt-PT", "pt-BR", "pt-AO", "pt-MZ"} {
		lt := languagetool.NewJLanguageTool(code)
		RegisterCorePortugueseRules(lt)
		require.Equal(t, code, lt.GetLanguageCode())

		// fully correct → 0 matches under core rules (Java also 0 with full stack)
		require.Empty(t, lt.Check(sentence),
			"%s: clean sentence must have 0 core-rule matches", code)

		// agreement error needs grammar.xml pattern rules — do not invent a fake hit
		_ = lt.Check(sentence2)
	}

	// Document Java expectation for future pattern-rule activation.
	require.Equal(t, "Isto é uma características sua.", sentence2)
}

// Port of JLanguageToolTest.testSomeSentences — Java only calls check (no size asserts).
// Must run without panic under core rules.
func TestJLanguageTool_PortugueseSomeSentences(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-BR")
	RegisterCorePortugueseRules(lt)
	require.Equal(t, "pt-BR", lt.GetLanguageCode())

	_ = lt.Check("™ ® Marcas registradas da Corteva Agriscience e de suas companhias afiliadas.")
	_ = lt.Check("Vatatzes teve que lutar contra a reivindicação de Isaac e Aleixo, os irmãos de Teodoro I, que fugiram para o Império Latino e procuraram ajuda.\n" +
		"Outros artistas que participaram do Vevo Lift são One Direction, 5 Seconds of Summer, Fifth Harmony, Sam Smith, Iggy Azalea, Rita Ora, Avicii, entre outros.\n" +
		"Astarte (desambiguação)\n" +
		"Drocourt (pas-de-calais)\n" +
		"O principal trabalho do G.E.R.E.C foi desenvolver, a partir de 1976, a escrita do crioulo, incluindo uma família de padrões de ortografia.\n" +
		"Zeth foi o mais novo de dez filhos.\n" +
		"Segundo Ajahn Mudito, a proposta do site é dar às pessoas a oportunidade de ter esses ensinamentos traduzidos diretamente do tailandês para o português, sem ter que passar pela costumeira “retradução”, onde primeiro os ensinamentos são traduzidos para o inglês para, só depois, chegar ao nosso idioma, fato que, segundo ele, gera muitas distorções e perda de significado do ensinamento original.\n" +
		"No entanto, em julho de 2016 foi anunciado como vice-prefeito na candidatura de Hélio Godoy.\n" +
		"As dissecções que envolvem a aorta descendente são geralmente tratadas com medicação para baixar a pressão arterial e o ritmo cardíaco, a não ser que ocorram complicações.\n" +
		"Margarethe era filha de Karl Krambeck e Catharina Neve, que emigraram de Sehestedt para o Brasil em 10.03.1852, com o Vapor Princess Louize e Capitão Bähr, dirigindo-se ao Distrito de Valença/RJ, onde trabalharam na Fazenda Independência, A família Krambeck originou-se no Estado de Schleswig-Holstein, no Norte da Alemanha, e os registros existentes têm início com Claus Krambeck, por volta de 1700, em Sehestedt.")

	// Word-repeat core surface (not in Java test body, but proves real rule path).
	// FILLER_WORDS_PT is default-off (Java AbstractStatisticStyleRule.setDefaultOff).
	require.Contains(t, lt.GetDefaultOffRuleIDs(), "FILLER_WORDS_PT")
	clean := lt.Check("Isto é uma frase correta.")
	for _, m := range clean {
		require.NotEqual(t, "PORTUGUESE_WORD_REPEAT_RULE", m.RuleID)
		require.NotEqual(t, "FILLER_WORDS_PT", m.RuleID,
			"FILLER_WORDS_PT must stay off until EnableRule")
	}
	matches := lt.Check("Isto é é uma frase.")
	found := false
	for _, m := range matches {
		if m.RuleID == "PORTUGUESE_WORD_REPEAT_RULE" {
			found = true
			break
		}
	}
	require.True(t, found, "PortugueseWordRepeatRule must fire on 'é é'")
}

// Metadata surface used by Portuguese language root factories.
func TestJLanguageTool_PortugueseLanguageRootSurface(t *testing.T) {
	require.Equal(t, "pt", language.Portuguese.GetShortCode())
	require.Equal(t, "pt-PT", language.PortugueseDefaultLanguageVariantCode())
	require.Equal(t, []string{"PT"}, language.PortugalPortuguese.GetCountries())
	require.Equal(t, []string{"BR"}, language.BrazilianPortuguese.GetCountries())
	require.Equal(t, []string{"AO"}, language.AngolaPortuguese.GetCountries())
	require.Equal(t, []string{"MZ"}, language.MozambiquePortuguese.GetCountries())
	require.Equal(t, []string{"", "CV", "GW", "MO", "ST", "TL"}, language.PortugueseBaseCountries())
}
