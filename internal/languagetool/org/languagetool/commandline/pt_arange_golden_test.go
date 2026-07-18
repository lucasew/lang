package commandline

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// PT soft disambig A_RANGE requires Z.+ around "a". Soft untagged must not treat
// letter words as numbers (Java Z = numeral), or every "a" becomes SP000 and
// blocks soft grammar postag matching under the hybrid disambiguator.
func TestGolden_PT_ARangeDoesNotTagBareA(t *testing.T) {
	lt, err := configureCoreLT("pt", &CommandLineOptions{Language: "pt"})
	require.NoError(t, err)
	// tags on "a" must not be SP000 from false A_RANGE
	for _, text := range []string{
		"É preciso afear a faca.",
		"A equação vai causar dificuldades a criar o algoritmo.",
	} {
		sents := lt.Analyze(text)
		require.NotEmpty(t, sents)
		for _, tok := range sents[0].GetTokensWithoutWhitespace() {
			if tok == nil || tok.GetToken() != "a" {
				continue
			}
			for _, r := range tok.GetReadings() {
				if r != nil && r.GetPOSTag() != nil {
					require.NotEqual(t, "SP000", *r.GetPOSTag(), "text=%q", text)
				}
			}
		}
	}
	require.True(t, checkHasRule(lt, "AFEAR", "É preciso afear a faca."))
	require.True(t, checkHasRule(lt, "CAUSAR_DIFICULDADE_DIFICULTAR",
		"A equação vai causar dificuldades a criar o algoritmo."))
	require.True(t, checkHasRule(lt, "ULTRAPASSAR_SUPERAR_VENCER", "A Ana vai ultrapassar a prova."))
}

func checkHasRule(lt *languagetool.JLanguageTool, rule, text string) bool {
	for _, m := range lt.Check(text) {
		if m.RuleID == rule {
			return true
		}
	}
	return false
}
