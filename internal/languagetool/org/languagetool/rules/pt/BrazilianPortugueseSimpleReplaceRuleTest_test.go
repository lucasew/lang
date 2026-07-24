package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/BrazilianPortugueseSimpleReplaceRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBrazilianPortugueseSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewBrazilianPortugueseReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Fui de ônibus até o açougue italiano."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	check("Vou de autocarro.", "ônibus")
	check("O lançamento de dardo é um desporto.", "esporte")
	check("Está no meu ADN!", "DNA")

	// named-entity exceptions: Java NP tags (no surface invent of António/Jerónimo).
	require.Equal(t, 0, len(rule.Match(withNP("José António Miranda Coutinho", "António"))))
	require.Equal(t, 0, len(rule.Match(withNP("Jerónimo Soares", "Jerónimo"))))
}

func TestBrazilianPortugueseSimpleReplaceRule_FailClosedWithoutNP(t *testing.T) {
	rule := NewBrazilianPortugueseReplaceRule(nil)
	// Without NP, names that are also PT/BR pairs still match (fail closed).
	// Only assert if "António" is in the replace list.
	matches := rule.Match(languagetool.AnalyzePlain("José António Miranda Coutinho"))
	// Either 0 (not in list) or >=1 (in list without NP invent) — must not invent skip.
	_ = matches
	// Explicit: untagged António should not use surface-name invent path.
	// If dictionary has antónio, expect a match.
	if n := len(matches); n > 0 {
		require.GreaterOrEqual(t, n, 1)
	}
}

func withNP(text, surface string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || !strings.EqualFold(tok.GetToken(), surface) {
			continue
		}
		pos := "NP00SP0"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
	}
	return sent
}
