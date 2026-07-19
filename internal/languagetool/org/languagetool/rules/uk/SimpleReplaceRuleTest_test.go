package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ці рядки повинні збігатися."))))

	matches := rule.Match(languagetool.AnalyzePlain("Ці рядки повинні співпадати"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"збігатися", "сходитися"}, matches[0].GetSuggestedReplacements())
	require.False(t, strings.Contains(matches[0].GetMessage(), "просторічна форма"))

	// Dictionary forms (untagged)
	matches = rule.Match(languagetool.AnalyzePlain("Нападаючий"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Нападник", "Нападальний", "Нападний"}, matches[0].GetSuggestedReplacements())

	// Java: lemma "нападаючий" in replace.txt → dictionary path (checkLemmas true), not adjp message.
	matches = rule.Match(analyzeUKTagged("Нападаючого", map[string][]languagetool.TokenTag{
		"Нападаючого": {{POS: "adjp:actv:m:v_rod:bad", Lemma: "нападаючий"}},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Нападник", "Нападальний", "Нападний"}, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "помилкове слово")

	// adjp:actv:bad only when super.findMatches empty (lemma not in replace.txt)
	matches = rule.Match(analyzeUKTagged("роблячий", map[string][]languagetool.TokenTag{
		"роблячий": {{POS: "adjp:actv:m:v_naz:bad", Lemma: "роблячий"}},
	}))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "Активні дієприкметники")

	// ignoreTagged: good POS skips (щедроти may be properly tagged in Java)
	// untagged щедрота matches dictionary
	matches = rule.Match(languagetool.AnalyzePlain("щедрота"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"щедрість", "гойність", "щедриня"}, matches[0].GetSuggestedReplacements())

	// good POS → ignore
	matches = rule.Match(analyzeUKTagged("щедроти", map[string][]languagetool.TokenTag{
		"щедроти": {{POS: "noun:p:v_naz", Lemma: "щедрота"}},
	}))
	require.Equal(t, 0, len(matches))
}

func TestSimpleReplaceRule_Derivat(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	// Surface in replace.txt (dictionary path)
	matches := rule.Match(languagetool.AnalyzePlain("перелиставши."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"перегорнувши", "прогортаючи"}, matches[0].GetSuggestedReplacements())
}

func TestSimpleReplaceRule_RulePartOfMultiword(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	// "проводжаючих" may need POS; dictionary or adjp path
	matches := rule.Match(languagetool.AnalyzePlain("на думку проводжаючих"))
	// Without tags: match if in replace.txt
	if len(matches) == 0 {
		matches = rule.Match(analyzeUKTagged("на думку проводжаючих", map[string][]languagetool.TokenTag{
			"проводжаючих": {{POS: "adjp:actv:p:v_rod:bad", Lemma: "проводжаючий"}},
		}))
	}
	require.Equal(t, 1, len(matches))
}

func TestSimpleReplaceRule_Misspellings(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	rule.SpellingSuggestions = func(word string) []string {
		if word == "ганделик" {
			return []string{"генделик", "ган делик"}
		}
		return nil
	}
	// Java: :bad POS + speller; without POS fail closed for this branch
	matches := rule.Match(analyzeUKTagged("ганделик", map[string][]languagetool.TokenTag{
		"ганделик": {{POS: "noun:m:v_naz:bad", Lemma: "ганделик"}},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"генделик"}, matches[0].GetSuggestedReplacements()) // space form filtered
	require.Equal(t, "Неправильно написане слово.", matches[0].GetMessage())
}

func TestSimpleReplaceRule_RuleByTag(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	// Java tags these as adjp:actv:…:bad
	for _, s := range []string{"спороутворюючого", "примкнувшим"} {
		matches := rule.Match(analyzeUKTagged(s, map[string][]languagetool.TokenTag{
			s: {{POS: "adjp:actv:m:v_rod:bad", Lemma: s}},
		}))
		require.Equal(t, 1, len(matches), s)
		require.Contains(t, matches[0].GetMessage(), "Активні дієприкметники", s)
	}
}

func TestSimpleReplaceRule_IgnoreTaggedWords(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Ці рядки повинні співпадати"))
	require.Equal(t, 1, len(matches))

	sent := languagetool.AnalyzeWithTagger("Ці рядки повинні співпадати", func(tok string) []languagetool.TokenTag {
		if tok == "співпадати" {
			return []languagetool.TokenTag{{POS: "verb:inf", Lemma: "співпадати"}}
		}
		return nil
	})
	require.Empty(t, rule.Match(sent), "good POS should skip replace")

	sentBad := languagetool.AnalyzeWithTagger("Ці рядки повинні співпадати", func(tok string) []languagetool.TokenTag {
		if tok == "співпадати" {
			return []languagetool.TokenTag{{POS: "verb:inf:bad", Lemma: "співпадати"}}
		}
		return nil
	})
	require.NotEmpty(t, rule.Match(sentBad), ":bad is not good POS for ignore")
}

func TestUkIsGoodPosTag(t *testing.T) {
	require.True(t, ukIsGoodPosTag("verb:inf"))
	require.False(t, ukIsGoodPosTag(""))
	require.False(t, ukIsGoodPosTag(languagetool.SentenceEndTagName))
	require.False(t, ukIsGoodPosTag("adj:bad"))
	require.False(t, ukIsGoodPosTag("noun:subst"))
	require.False(t, ukIsGoodPosTag("<ignore>"))
}

func TestFindInDeriv(t *testing.T) {
	// If surface is only in derivats as key and base verb is in replace.txt
	// Smoke: load succeeds and known key exists
	m := loadDerivats()
	require.NotEmpty(t, m)
	_, ok := m["перелиставши"]
	require.True(t, ok, "перелиставши should be in derivats")
}

func analyzeUKTagged(text string, tags map[string][]languagetool.TokenTag) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		if t, ok := tags[tok]; ok {
			return t
		}
		// also try lower
		if t, ok := tags[strings.ToLower(tok)]; ok {
			return t
		}
		return nil
	})
}
