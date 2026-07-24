package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianWordRepeatRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianWordRepeatRule_Rule(t *testing.T) {
	rule := NewRussianWordRepeatRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Повтор слов в предложении."))))
	// Java tagger lowercases lemmas (Повтор/повтор → повтор). Inject lemmas — no ToLower invent.
	require.Equal(t, 1, len(rule.Match(withLemma("Повтор слов в повтор предложении.", "повтор", "повтор"))))
}

func withLemma(text, surface, lemma string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || !strings.EqualFold(tok.GetToken(), surface) {
			continue
		}
		pos := "NN"
		lem := lemma
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, &lem), "test")
	}
	return sent
}
