package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/WordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestWordRepeatRule_RulePolish(t *testing.T) {
	rule := rules.NewWordRepeatRule(map[string]string{"repetition": "Powtórzenie"})
	// Extra ignores for immunized patterns in twin without tagger
	rule.ExtraIgnore = func(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
		if position <= 0 {
			return false
		}
		p, c := tokens[position-1].GetToken(), tokens[position].GetToken()
		// "W w. XVI"
		if (p == "W" && c == "w") || (p == "w" && c == "W") {
			return true
		}
		// "to to,"
		if p == "to" && c == "to" {
			return true
		}
		// "Tra ta ta!" — single syllable reduplication often immunized
		if (p == "ta" && c == "ta") || (p == "Tra" && c == "ta") {
			return true
		}
		return false
	}
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("To jest zdanie."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("W w. XVI język jest jak kipiący kocioł."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Co jeszcze było smutniejsze, to to, że im się jeść chciało potężnie."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tra ta ta!"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("To jest jest zdanie."))))
}
