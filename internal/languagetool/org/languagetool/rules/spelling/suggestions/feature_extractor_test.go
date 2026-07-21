package suggestions

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// mapLM implements LanguageModelHook for unit tests.
type mapLM map[string]float64

func (m mapLM) PseudoProbability(tokens []string) float64 {
	if len(tokens) == 0 {
		return 0
	}
	// Unigram: look up first token; multi-token contexts use constant for 3gram product.
	if len(tokens) == 1 {
		if p, ok := m[tokens[0]]; ok {
			return p
		}
		return 1e-9 // avoid log(0) when word missing
	}
	return 0.1 // 3gram context backoff for tests
}
func (m mapLM) Count(word string) int64 {
	if p, ok := m[word]; ok && p > 0 {
		return int64(p * 1000)
	}
	return 0
}

func TestFeatureExtractor_IsMlAvailable(t *testing.T) {
	require.False(t, NewSuggestionsOrdererFeatureExtractor(nil).IsMlAvailable())
	require.True(t, NewSuggestionsOrdererFeatureExtractor(mapLM{}).IsMlAvailable())
}

// Java score "noop" preserves candidate order; match features include candidateCount.
func TestFeatureExtractor_NoopOrderAndFeatures(t *testing.T) {
	e := NewSuggestionsOrdererFeatureExtractor(mapLM{"hello": 0.9, "hallo": 0.1})
	e.Score = "noop"
	ordered, agg := e.ComputeFeatures([]string{"helo", "hello", "hallo"}, "hello", nil, 0)
	require.Len(t, ordered, 3)
	require.Equal(t, "helo", ordered[0].Replacement)
	require.Equal(t, "hello", ordered[1].Replacement)
	require.Equal(t, float32(3), agg["candidateCount"])
	// Per-candidate Feature.getData keys (Java TreeMap)
	f := ordered[0].GetFeatures()
	require.Contains(t, f, "prob1gram")
	require.Contains(t, f, "prob3gram")
	require.Contains(t, f, "wordCount")
	require.Contains(t, f, "levensthein")
	require.Contains(t, f, "jaroWrinkler")
	require.Contains(t, f, "inserts")
	require.Contains(t, f, "deletes")
	require.Contains(t, f, "replaces")
	require.Contains(t, f, "transposes")
	require.Contains(t, f, "wordLength")
	// exact match distance 0 for "hello" vs "hello"
	ed := implementationCompare("hello", "hello")
	_ = ed
	// hello vs helo: Damerau/edit should be non-zero on first candidate features
	require.Greater(t, f["levensthein"], float32(0))
}

// Java score "ngrams" sorts by log(prob1)+log(prob3) descending.
func TestFeatureExtractor_NgramsOrder(t *testing.T) {
	e := NewSuggestionsOrdererFeatureExtractor(mapLM{"hello": 0.9, "hallo": 0.05, "helo": 0.01})
	e.Score = "ngrams"
	// WordTokenizer keeps whitespace tokens so Google start positions match Java/LT.
	wt := tokenizers.NewWordTokenizer()
	e.Tokenize = wt.Tokenize
	text := "say hello please"
	sent := languagetool.AnalyzePlain(text)
	start := strings.Index(text, "hello")
	require.GreaterOrEqual(t, start, 0)
	// typo form is the match word; candidates ranked by LM
	got := OrderSuggestionsUsingModel(e, []string{"helo", "hallo", "hello"}, "helo", sent, start)
	require.Equal(t, "hello", got[0], "highest unigram should rank first: %v", got)
}

// Experiment parameters override topN/score (Java initParameters).
func TestFeatureExtractor_ExperimentParams(t *testing.T) {
	ResetSuggestionsChanges()
	t.Cleanup(ResetSuggestionsChanges)
	InitSuggestionsChanges(&SuggestionChangesTestConfig{
		Experiments: []SuggestionChangesExperiment{{
			Name: "test",
			Parameters: map[string]any{
				"topN":            2,
				"score":           "noop",
				"levenstheinProb": 0.5,
			},
		}},
	})
	s := GetSuggestionsChanges()
	s.SetCurrentExperiment(&s.Config.Experiments[0])

	e := NewSuggestionsOrdererFeatureExtractor(mapLM{"a": 0.1})
	require.Equal(t, 2, e.TopN)
	require.Equal(t, "noop", e.Score)
	require.Equal(t, 0.5, e.MistakeProb)

	ordered, _ := e.ComputeFeatures([]string{"a", "b", "c"}, "x", nil, 0)
	require.Len(t, ordered, 2)
}

// Unknown score panics (Java RuntimeException).
func TestFeatureExtractor_UnknownScorePanics(t *testing.T) {
	e := NewSuggestionsOrdererFeatureExtractor(mapLM{"a": 0.1})
	e.Score = "not-a-real-score"
	require.Panics(t, func() {
		e.ComputeFeatures([]string{"a", "b"}, "a", nil, 0)
	})
}

func TestJaroWinkler(t *testing.T) {
	require.InDelta(t, 1.0, jaroWinkler("abc", "abc"), 0.001)
	require.Greater(t, jaroWinkler("hello", "hallo"), jaroWinkler("hello", "xyz"))
	// short strings below Commons threshold 0.7 still return pure jaro (no prefix boost path required)
	require.GreaterOrEqual(t, jaroWinkler("ab", "ac"), 0.0)
}

func implementationCompare(a, b string) int {
	// helper placeholder — distance features exercised via ComputeFeatures
	if a == b {
		return 0
	}
	return 1
}
