package eval

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/dev/errorcorpus"
	"github.com/stretchr/testify/require"
)

func TestRunSimpleCorpus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.txt")
	require.NoError(t, os.WriteFile(path, []byte("1. This is _a_ error. => an\n"), 0o644))
	c, err := errorcorpus.NewSimpleCorpus(path)
	require.NoError(t, err)
	ev := NewRealWordCorpusEvaluator(FuncEvaluator{Fn: func(text string) ([]Match, error) {
		// match "a" → "an"
		idx := 8 // "This is a error." → a at 8
		return []Match{{FromPos: idx, ToPos: idx + 1, SuggestedReplacements: []string{"an"}}}, nil
	}})
	require.NoError(t, ev.RunSimpleCorpus(c))
	require.Equal(t, 1, ev.SentenceCount)
	require.Equal(t, 1, ev.ErrorsInCorpusCount)
	require.Equal(t, 1, ev.GoodMatches)
	require.Equal(t, 1, ev.PerfectMatches)
}

func TestRunPedlerCorpus(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "p.txt"), []byte(
		"She <ERR targ=saw>see</ERR> it.\n",
	), 0o644))
	c, err := errorcorpus.NewPedlerCorpus(dir)
	require.NoError(t, err)
	ev := NewRealWordCorpusEvaluator(FuncEvaluator{Fn: func(text string) ([]Match, error) {
		// plain: "She see it."
		i := 4
		return []Match{{FromPos: i, ToPos: i + 3, SuggestedReplacements: []string{"saw"}}}, nil
	}})
	require.NoError(t, ev.RunPedlerCorpus(c))
	require.Equal(t, 1, ev.PerfectMatches)
}
