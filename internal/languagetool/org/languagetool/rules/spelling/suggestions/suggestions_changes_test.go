package suggestions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionsChanges_IsRunningExperiment(t *testing.T) {
	t.Cleanup(ResetSuggestionsChanges)
	require.False(t, IsRunningExperiment("x"))

	s := InitSuggestionsChanges(&SuggestionChangesTestConfig{Language: "en"})
	require.Same(t, s, GetSuggestionsChanges())
	e := &SuggestionChangesExperiment{Name: "A", Parameters: map[string]any{"topN": 5}}
	s.SetCurrentExperiment(e)
	require.True(t, IsRunningExperiment("A"))
	require.False(t, IsRunningExperiment("B"))
}

// Twin of trackExperimentResult: pos 0 correct, -1 not found, else pos sum.
func TestSuggestionsChanges_TrackExperimentResult(t *testing.T) {
	t.Cleanup(ResetSuggestionsChanges)
	s := InitSuggestionsChanges(&SuggestionChangesTestConfig{})
	e := &SuggestionChangesExperiment{Name: "A", Parameters: map[string]any{"topN": 5}}
	s.SetCurrentExperiment(e)

	// Java convenience path used by older Go tests
	s.RecordCorrect()
	s.RecordNotFound()
	s.RecordSuggestionPos(2)

	require.Equal(t, 1, s.CorrectCount(e))
	require.Equal(t, 1, s.NotFoundCount(e))
	require.Equal(t, 2, s.PosSum(e)) // 0 + 2 (not-found does not add to pos sum)
	require.Equal(t, 3, s.NumSamples(e))
}

// gridsearch: TreeMap lastKey peel — two dimensions expand to cartesian product.
func TestSuggestionsChanges_GenerateExperimentsGrid(t *testing.T) {
	t.Cleanup(ResetSuggestionsChanges)
	s := InitSuggestionsChanges(&SuggestionChangesTestConfig{
		ExperimentRuns: []SuggestionChangesExperimentRuns{
			{
				Name: "grid",
				Parameters: map[string][]any{
					"score": {"ngrams", "noop"},
					"topN":  {1, 2},
				},
			},
			{Name: "plain"}, // nil params → one empty experiment
		},
	})
	exps := s.GetExperiments()
	// 2*2 + 1 = 5
	require.Len(t, exps, 5)
	// all named grid have both keys
	var gridCount int
	for _, e := range exps {
		if e.Name == "plain" {
			require.Empty(t, e.Parameters)
			continue
		}
		require.Equal(t, "grid", e.Name)
		require.Contains(t, e.Parameters, "score")
		require.Contains(t, e.Parameters, "topN")
		gridCount++
	}
	require.Equal(t, 4, gridCount)
}

// Dataset-scoped counters + report text (Java Report.run format fragments).
func TestSuggestionsChanges_DatasetAndReport(t *testing.T) {
	t.Cleanup(ResetSuggestionsChanges)
	ds := SuggestionChangesDataset{Name: "corpus1", Type: "artificial"}
	s := InitSuggestionsChanges(&SuggestionChangesTestConfig{
		ExperimentRuns: []SuggestionChangesExperimentRuns{
			{Name: "expA", Parameters: map[string][]any{"k": {"v"}}},
			{Name: "expB"},
		},
		Datasets: []SuggestionChangesDataset{ds},
	})
	exps := s.GetExperiments()
	require.Len(t, exps, 2)

	// expA: 2 correct of 3, positions 0,0,1
	s.TrackExperimentResult(exps[0], &ds, 0, 10, 5)
	s.TrackExperimentResult(exps[0], &ds, 0, 10, 5)
	s.TrackExperimentResult(exps[0], &ds, 1, 10, 5)
	// expB: not found
	s.TrackExperimentResult(exps[1], &ds, -1, 5, 2)

	require.Equal(t, 2, s.CorrectCount(exps[0]))
	require.Equal(t, 3, s.NumSamples(exps[0]))
	require.Equal(t, 1, s.NotFoundCount(exps[1]))

	report := s.BuildReport()
	require.Contains(t, report, "Overall report:")
	require.Contains(t, report, "Experiment #1")
	require.Contains(t, report, "Experiment #2")
	require.Contains(t, report, "Best experiment:")
	require.Contains(t, report, "Report for dataset: corpus1")
	// expA should win accuracy ~66%
	require.True(t, strings.Contains(report, "expA") || strings.Contains(report, "name=expA"))
}

// Java: position 0 also adds to suggestionPosSum (else branch), so score includes 0s.
func TestSuggestionsChanges_PosZeroInSum(t *testing.T) {
	t.Cleanup(ResetSuggestionsChanges)
	s := InitSuggestionsChanges(&SuggestionChangesTestConfig{})
	e := &SuggestionChangesExperiment{Name: "z", Parameters: map[string]any{}}
	s.TrackExperimentResult(e, nil, 0, 0, 0)
	s.TrackExperimentResult(e, nil, 3, 0, 0)
	require.Equal(t, 1, s.CorrectCount(e))
	require.Equal(t, 3, s.PosSum(e)) // 0+3
	require.Equal(t, 2, s.NumSamples(e))
}

func TestGridsearch_OrderMatchesTreeMap(t *testing.T) {
	// Single dimension
	got := gridsearch(map[string][]any{"a": {1, 2}}, nil)
	require.Len(t, got, 2)
	require.Equal(t, 1, got[0]["a"])
	require.Equal(t, 2, got[1]["a"])

	// Two dimensions: lastKey highest first — combinations product size
	got = gridsearch(map[string][]any{
		"a": {1, 2},
		"b": {"x", "y"},
	}, nil)
	require.Len(t, got, 4)
	// every pair present
	type pair struct {
		a int
		b string
	}
	seen := map[pair]bool{}
	for _, m := range got {
		seen[pair{a: m["a"].(int), b: m["b"].(string)}] = true
	}
	require.True(t, seen[pair{1, "x"}])
	require.True(t, seen[pair{1, "y"}])
	require.True(t, seen[pair{2, "x"}])
	require.True(t, seen[pair{2, "y"}])
}
