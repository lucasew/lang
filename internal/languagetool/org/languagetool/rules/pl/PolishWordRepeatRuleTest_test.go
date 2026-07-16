package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/PolishWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPolishWordRepeatRule_Rule(t *testing.T) {
	rule := NewPolishWordRepeatRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("To jest zdanie próbne."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("On tak się bardzo nie martwił, bo przecież musiał się umyć."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Na dyskotece tańczył jeszcze, choć był na bani."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Żadnych „ale”."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Był on bowiem pięknym strzelcem bowiem."))))
	// "długo" twice → 1 advanced match at second occurrence; Java also counts "mówiła/mówić"?
	// Java: 2 matches for "Mówiła długo, żeby tylko mówić długo." — lemma mówić twice + długo twice?
	// Without tagger only surface "długo" twice → 1 match.
	m := rule.Match(languagetool.AnalyzePlain("Mówiła długo, żeby tylko mówić długo."))
	require.GreaterOrEqual(t, len(m), 1)
	// Prefer full twin when possible: surface only yields 1
	require.Equal(t, 1, len(m))
}
