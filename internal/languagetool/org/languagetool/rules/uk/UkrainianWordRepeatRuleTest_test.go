package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/UkrainianWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkrainianWordRepeatRule_Rule(t *testing.T) {
	rule := NewUkrainianWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	ok := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "ok %q", s)
	}
	ok("без повного розрахунку")
	ok("без бугіма бугіма") // may need redup allow
	ok("без 100 100")
	ok("1.30 3.20 3.20")
	ok("ще в В.Кандинського")
	ok("Від добра добра не шукають.")
	ok("Що, що, а кіно в Україні...")
	ok("Відповідно до ст. ст. 3, 7, 18.")
	ok("Не можу сказати ні так, ні ні.")

	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("без без повного розрахунку"))))
	match := rule.Match(languagetool.AnalyzePlain("Верховної Ради І і ІІ скликань"))
	require.Equal(t, 1, len(match))
	require.Equal(t, 2, len(match[0].GetSuggestedReplacements()))
}
