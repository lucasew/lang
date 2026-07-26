package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseUnitConversionRuleTest.java
// Helper mirrors languagetool-core/.../UnitConversionRuleTestHelper.assertMatches:
// match count + first match's suggestions must contain the converted substring.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// assertPTUnitConversionMatches ports UnitConversionRuleTestHelper.assertMatches
// (JLanguageTool.getAnalyzedSentence → rule.match; count + suggestion contains).
func assertPTUnitConversionMatches(t *testing.T, rule *PortugueseUnitConversionRule, input string, expectedMatches int, converted string) {
	t.Helper()
	// Java: lt.getAnalyzedSentence(input) — plain analysis (no POS required by this rule).
	ms := rule.Match(languagetool.AnalyzePlain(input))
	require.Equalf(t, expectedMatches, len(ms), "Got matches for %q: %v", input, summarizePTUnitMatches(ms))
	if expectedMatches > 0 && converted != "" {
		require.NotEmpty(t, ms)
		suggestionCorrect := false
		var suggestion string
		for _, s := range ms[0].GetSuggestedReplacements() {
			if strings.Contains(s, converted) {
				suggestionCorrect = true
				suggestion = s
				break
			}
		}
		require.Truef(t, suggestionCorrect, "Suggestion is correct: %q / expected contains: %q (all: %v)",
			suggestion, converted, ms[0].GetSuggestedReplacements())
	}
}

func summarizePTUnitMatches(ms []*rules.RuleMatch) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		if m == nil {
			out = append(out, "<nil>")
			continue
		}
		out = append(out, m.GetMessage()+" | "+strings.Join(m.GetSuggestedReplacements(), "; "))
	}
	return out
}

func TestPortugueseUnitConversionRule_Match(t *testing.T) {
	// Java: Language lang = Languages.getLanguageForShortCode("pt");
	//       JLanguageTool lt = new JLanguageTool(lang);
	//       PortugueseUnitConversionRule rule = new PortugueseUnitConversionRule(...);
	rule := NewPortugueseUnitConversionRule(nil)
	require.Equal(t, "UNIDADES_METRICAS", rule.GetID())
	require.Equal(t, "Sugere ou verifica informações equivalentes à métrica de unidades de medida específicas.", rule.GetDescription())
	// Java formatRounded
	require.Equal(t, "aprox. 1,83 m", rule.FormatRounded("1,83 m"))

	// Full Java PortugueseUnitConversionRuleTest.match() matrix
	assertPTUnitConversionMatches(t, rule, "Eu tenho 6 pés de altura.", 1, "1,83 metros")
	assertPTUnitConversionMatches(t, rule, "Eu tenho 6 pés (2,02 m) de altura.", 1, "1,83 metros")
	assertPTUnitConversionMatches(t, rule, "Eu tenho 6 pés (1,82 m) de altura.", 0, "")
	assertPTUnitConversionMatches(t, rule, "A via tem 100 milhas de comprimento.", 1, "160,93 quilômetros")
	assertPTUnitConversionMatches(t, rule, "A via tem 10 km (20 milhas) de comprimento.", 1, "6,21")
	assertPTUnitConversionMatches(t, rule, "A via tem 10 km (6,21 milhas) de comprimento.", 0, "")
	assertPTUnitConversionMatches(t, rule, "A via tem 100 milhas (160,93 quilômetros) de comprimento.", 0, "")
	assertPTUnitConversionMatches(t, rule, "A carga é de 10.000 libras.", 1, "4,54 toneladas")
	assertPTUnitConversionMatches(t, rule, `Isto tem 5'6" de altura.`, 1, "1,68 m")
	assertPTUnitConversionMatches(t, rule, "O meu novo apartamento tem 500 sq ft de área.", 1, "46,45 metros quadrados")
	assertPTUnitConversionMatches(t, rule, `Sendo a latitude 8º 32' 00" e a longitude 39º 22' 49".`, 0, "")
	assertPTUnitConversionMatches(t, rule, `Sendo a latitude 8º32'00" e a longitude 39º22'49".`, 0, "")
	// Same unit, assume something shady, but don't suggest conversion to the same one!
	assertPTUnitConversionMatches(t, rule, "Com altura de 4,8 cm (22,08 cm).", 0, "")
	// Composite expression (squaring); should work properly at some point, but at least now it doesn't fail
	assertPTUnitConversionMatches(t, rule, "Com altura de 4,8 mm x 5,2 cm (2,5 cm²).", 0, "")
	assertPTUnitConversionMatches(t, rule, "Com altura de 4,8 mm × 5,2 cm (2,5 cm²).", 0, "")
}
