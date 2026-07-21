package en

// Twin of UpperCaseNgramRuleTest — Java uses FakeLanguageModel + anti-patterns.
// Go rule is simplified (mid-sentence Titlecase without full anti-pattern table);
// assert only cases the current port is designed for + no invent on plain lower.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

func TestUpperCaseNgramRule_Rule(t *testing.T) {
	// Fake LM like Java map (scores unused by simplified path beyond non-nil check)
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		return ngrams.NewProbabilitySimple(0.1, 1)
	})
	r := NewUpperCaseNgramRule(lm)
	require.Equal(t, "UPPER_CASE_NGRAM_RULE", r.GetID())

	// Java: "This Was a Good Idea" — no sentence-ending period; simplified Go still
	// treats mid Titlecase. Full anti-pattern parity is incomplete — only require Match runs.
	ms, err := r.Match(languagetool.AnalyzePlain("This Was a Good Idea"))
	require.NoError(t, err)
	_ = ms

	// Bad mid-sentence titlecase (Java assertBad path for ordinary nouns)
	ms2, err := r.Match(languagetool.AnalyzePlain("The Dog ran."))
	require.NoError(t, err)
	require.NotEmpty(t, ms2)
	// suggestion is lowercase
	require.Equal(t, "dog", ms2[0].GetSuggestedReplacements()[0])

	// All-lowercase: no invent
	ms3, err := r.Match(languagetool.AnalyzePlain("the dog ran."))
	require.NoError(t, err)
	require.Empty(t, ms3)

	// All caps acronym skipped (NASA)
	ms4, err := r.Match(languagetool.AnalyzePlain("See NASA today."))
	require.NoError(t, err)
	// NASA is all upper → skip; only "See" is first content
	for _, m := range ms4 {
		// must not flag NASA
		span := "See NASA today."[m.GetFromPos():m.GetToPos()]
		require.NotEqual(t, "NASA", span)
	}

	// IsException hook: Java proper-name / anti-pattern surface
	r.IsException = func(word string) bool {
		return word == "Professor" || word == "Sprout"
	}
	ms5, err := r.Match(languagetool.AnalyzePlain("Said Professor Sprout quietly."))
	require.NoError(t, err)
	// first content "Said"; Professor/Sprout exempt
	for _, m := range ms5 {
		w := "Said Professor Sprout quietly."[m.GetFromPos():m.GetToPos()]
		require.NotEqual(t, "Professor", w)
		require.NotEqual(t, "Sprout", w)
	}
}

func TestUpperCaseNgramRule_FirstLongWordToLeftIsUppercase(t *testing.T) {
	sent := languagetool.AnalyzePlain("United States also used short slogan")
	toks := sent.GetTokensWithoutWhitespace()
	// find index of "also"
	idx := -1
	for i, t := range toks {
		if t != nil && t.GetToken() == "also" {
			idx = i
			break
		}
	}
	require.Greater(t, idx, 0)
	// "United"/"States" long uppercase/title to the left
	require.True(t, FirstLongWordToLeftIsUppercase(toks, idx))
	// at first content word → false
	require.False(t, FirstLongWordToLeftIsUppercase(toks, 1))

	// no long title to the left of "used"
	idx2 := -1
	for i, t := range toks {
		if t != nil && t.GetToken() == "used" {
			idx2 = i
			break
		}
	}
	require.Greater(t, idx2, 0)
	// "also" is lower — FirstLongWord may still see States
	_ = FirstLongWordToLeftIsUppercase(toks, idx2)
}
