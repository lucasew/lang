package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseWordRepeatBeginningRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func ptWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning_adv":       "Três frases sucessivas começam com o mesmo advérbio.",
		"desc_repetition_beginning_word":      "Três frases sucessivas começam com a mesma palavra.",
		"desc_repetition_beginning_thesaurus": "Considere usar um dicionário de sinónimos.",
	}
}

func TestPortugueseWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewPortugueseWordRepeatBeginningRule(ptWRBMessages())

	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Este exemplo está correto. Este exemplo também está."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("2011: Setembro já passou. 2011: Outubro também já passou. 2011: Novembro já se foi."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Certo, isto está bem. Este exemplo está correto. Certo que este também."))))
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Este exemplo está correto. Este segundo também. Este terceiro exemplo não."))))
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Então, este está correto. Então, este está errado, por causa da repetição."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Então, este deve ser considerado uma nova frase."))))
}
