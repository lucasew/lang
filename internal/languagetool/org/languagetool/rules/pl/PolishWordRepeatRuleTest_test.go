package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/PolishWordRepeatRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPolishWordRepeatRule_Rule(t *testing.T) {
	rule := NewPolishWordRepeatRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("To jest zdanie próbne."))))
	// się twice: Java EXC_WORDS lemma "się" — inject lemma (no surface invent of prep list).
	require.Equal(t, 0, len(rule.Match(withLemma("On tak się bardzo nie martwił, bo przecież musiał się umyć.", "się", "się"))))
	// na twice: Java prep:.* POS — inject prep tag.
	require.Equal(t, 0, len(rule.Match(withPOS("Na dyskotece tańczył jeszcze, choć był na bani.", "na", "prep:loc:nwok"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Żadnych „ale”."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Był on bowiem pięknym strzelcem bowiem."))))
	// Without tagger only surface "długo" twice → 1 match.
	m := rule.Match(languagetool.AnalyzePlain("Mówiła długo, żeby tylko mówić długo."))
	require.Equal(t, 1, len(m))
}

func TestPolishWordRepeatRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewPolishWordRepeatRule(nil)
	// Without prep POS, same surface "na" twice is a style repeat (fail closed).
	// Java matches ExcludedWords/ExcludedPos on lemma/POS only — no surface invent.
	// Untagged path is case-sensitive (Java getToken()); use same-case surfaces.
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("tańczył na dyskotece choć był na bani."))))
}

func withLemma(text, surface, lemma string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || !strings.EqualFold(tok.GetToken(), surface) {
			continue
		}
		// Need non-empty POS for AdvancedWordRepeat lemma path (Java).
		pos := "qub"
		lem := lemma
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, &lem), "test")
	}
	return sent
}

func withPOS(text, surface, posTag string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || !strings.EqualFold(tok.GetToken(), surface) {
			continue
		}
		pos := posTag
		lem := strings.ToLower(tok.GetToken())
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, &lem), "test")
	}
	return sent
}
