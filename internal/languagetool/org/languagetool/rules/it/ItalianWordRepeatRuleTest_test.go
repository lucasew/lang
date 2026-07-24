package it

// Twin of languagetool-language-modules/it/src/test/java/org/languagetool/rules/it/ItalianWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestItalianWordRepeatRule_Rule(t *testing.T) {
	rule := NewItalianWordRepeatRule(map[string]string{"repetition": "Ripetizione"})
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Mi è sembrato così così"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Duran Duran"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Devi mescolare piano piano"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Seguo passo passo"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Mi mi è sembrato così"))))
}
