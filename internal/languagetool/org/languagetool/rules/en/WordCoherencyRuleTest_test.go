package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/WordCoherencyRuleTest.java
// Production loads coherency.txt pairs only (no invent -ed/-ing). Inflected forms
// match via lemmas (Java tagger); tests attach morph lemmas for twin inputs.
import (
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// twinCoherencyLemmas: surface → lemma for WordCoherencyRuleTest (Java morph).
var twinCoherencyLemmas = map[string]string{
	"archeology": "archeology", "archaeology": "archaeology",
	"reelect": "reelect", "re-elect": "re-elect",
	"reelected": "reelect", "re-elected": "re-elect",
	"oxidise": "oxidise", "oxidize": "oxidize",
	"oxidises": "oxidise", "oxidizes": "oxidize",
	"westernise": "westernise", "westernize": "westernize",
}

func analyzeCoherency(s string) []*languagetool.AnalyzedSentence {
	// Sentence-local positions; TextLevelRule adds GetCorrectedTextLength (Java analyzeText).
	if s == "" {
		return nil
	}
	var parts []string
	start := 0
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '.' || r == '!' || r == '?' {
			if r == '.' && i+1 < len(runes) {
				n := runes[i+1]
				if (n >= 'a' && n <= 'z') || (n >= '0' && n <= '9') {
					continue
				}
			}
			end := i + 1
			if end < len(runes) && (runes[end] == ' ' || runes[end] == '\n' || runes[end] == '\u00A0') {
				if runes[end] == '\n' && end+1 < len(runes) && runes[end+1] == '\n' {
					end++
					if end < len(runes) && runes[end] == '\n' {
						end++
					}
				} else if runes[end] == ' ' || runes[end] == '\u00A0' {
					end++
				} else if runes[end] == '\n' {
					end++
				}
			}
			parts = append(parts, string(runes[start:end]))
			start = end
			i = end - 1
		}
	}
	if start < len(runes) {
		parts = append(parts, string(runes[start:]))
	}
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, languagetool.AnalyzeWithTagger(p, enCoherencyTagWord))
	}
	return out
}

func enCoherencyTagWord(tok string) []languagetool.TokenTag {
	key := strings.ToLower(tok)
	key = strings.TrimFunc(key, func(r rune) bool { return !unicode.IsLetter(r) && r != '-' })
	if lem, ok := twinCoherencyLemmas[key]; ok {
		return []languagetool.TokenTag{{Lemma: lem}}
	}
	return nil
}

func TestWordCoherencyRule_Rule(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList(analyzeCoherency(s))), "good: %q", s)
	}
	assertError := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 1, len(rule.MatchList(analyzeCoherency(s))), "error: %q", s)
	}

	assertGood("He likes archeology. She likes archeology, too.")
	assertGood("He likes archaeology. She likes archaeology, too.")
	assertError("He likes archaeology. She likes archeology, too.")

	rule := NewWordCoherencyRule(nil)
	matches1 := rule.MatchList(analyzeCoherency("He is reelected, or he will be re-elected."))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 31, matches1[0].GetFromPos())
	require.Equal(t, 41, matches1[0].GetToPos())
	require.Equal(t, []string{"reelected"}, matches1[0].GetSuggestedReplacements())

	matches2 := rule.MatchList(analyzeCoherency("He was reelected, and I will re-elect him again in 2002."))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 29, matches2[0].GetFromPos())
	require.Equal(t, 37, matches2[0].GetToPos())
	require.Equal(t, []string{"reelect"}, matches2[0].GetSuggestedReplacements())

	matches3 := rule.MatchList(analyzeCoherency("He oxidises o, or he oxidizes"))
	require.Equal(t, 1, len(matches3))
	require.Equal(t, 21, matches3[0].GetFromPos())
	require.Equal(t, 29, matches3[0].GetToPos())
	require.Equal(t, []string{"oxidises"}, matches3[0].GetSuggestedReplacements())
}

func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	// Separate rule instances / match calls do not share state
	assertGood := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList(analyzeCoherency(s))))
	}
	assertGood("He likes archaeology.")
	assertGood("She likes archeology, too.")
}

func TestWordCoherencyRule_MatchPosition(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	matches := rule.MatchList(analyzeCoherency("He likes archaeology. She likes archeology, too."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 32, matches[0].GetFromPos())
	require.Equal(t, 42, matches[0].GetToPos())
}

func TestWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	// Java uses lt.check; we only run WordCoherencyRule
	check := func(s string) int {
		return len(NewWordCoherencyRule(nil).MatchList(analyzeCoherency(s)))
	}
	require.Equal(t, 0, check("He likes archaeology. Really? She likes archaeology, too."))
	require.Equal(t, 1, check("He likes archaeology. Really? She likes archeology, too."))
	require.Equal(t, 1, check("He likes archeology. Really? She likes archaeology, too."))
	require.Equal(t, 1, check("Mix of upper case and lower case: Westernize and westernise."))
}
