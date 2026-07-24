package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/UkrainianWordRepeatRuleTest.java
// POS inject for cases where Java Ukrainian tagger provides readings.
import (
	"strings"
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
	// untagged doubles: Java ignore() returns true (no non-initial POS)
	ok("без повного розрахунку")
	ok("без бугіма бугіма")
	ok("без 100 100")
	ok("1.30 3.20 3.20")
	ok("ще в В.Кандинського") // "В" + "." isInitial
	ok("Від добра добра не шукають.")
	ok("Що, що, а кіно в Україні...") // not adjacent tokens
	ok("Відповідно до ст. ст. 3, 7, 18.")
	ok("Не можу сказати ні так, ні ні.")

	// "без без" with prep POS → do not ignore (Java tagged prep)
	sent := languagetool.AnalyzeWithTagger("без без повного розрахунку", func(tok string) []languagetool.TokenTag {
		if strings.EqualFold(tok, "без") {
			return []languagetool.TokenTag{{POS: "prep", Lemma: "без"}}
		}
		return nil
	})
	require.Equal(t, 1, len(rule.Match(sent)))

	// Without POS, untagged "без без" is ignored (fail closed / Java no-tag path)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("без без повного розрахунку"))))

	// І і — letters with POS or untagged? "І" and "і" equalFold — without POS would ignore.
	// Java tags them; inject letter POS so match fires.
	sentI := languagetool.AnalyzeWithTagger("Верховної Ради І і ІІ скликань", func(tok string) []languagetool.TokenTag {
		if tok == "І" || tok == "і" {
			return []languagetool.TokenTag{{POS: "part", Lemma: tok}}
		}
		return nil
	})
	match := rule.Match(sentI)
	require.Equal(t, 1, len(match))
	require.Equal(t, 2, len(match[0].GetSuggestedReplacements()))
	require.Contains(t, match[0].GetSuggestedReplacements(), "I і")
}

func TestUkrainianWordRepeatRule_DateTimeNumPOS(t *testing.T) {
	rule := NewUkrainianWordRepeatRule(map[string]string{"repetition": "Повтор"})
	// number POS ignores doubles
	sent := languagetool.AnalyzeWithTagger("100 100", func(tok string) []languagetool.TokenTag {
		if tok == "100" {
			return []languagetool.TokenTag{{POS: "number", Lemma: "100"}}
		}
		return nil
	})
	// isWord("100") is false via IsNumericSpace — no match either way
	require.Equal(t, 0, len(rule.Match(sent)))

	// non-numeric surface with number.* POS
	sent2 := languagetool.AnalyzeWithTagger("foo foo", func(tok string) []languagetool.TokenTag {
		if tok == "foo" {
			return []languagetool.TokenTag{{POS: "number:latin", Lemma: "foo"}}
		}
		return nil
	})
	require.Equal(t, 0, len(rule.Match(sent2)), "number.* POS should ignore")
}
