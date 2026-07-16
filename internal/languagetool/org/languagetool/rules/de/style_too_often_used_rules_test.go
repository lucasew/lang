package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleTooOftenUsedVerbRule(t *testing.T) {
	rule := NewStyleTooOftenUsedVerbRule(nil)
	sents := languagetool.SplitAndAnalyze("Sie laufen schnell. Dann laufen sie weiter.")
	require.GreaterOrEqual(t, len(rule.MatchList(sents)), 2)
}

func TestStyleTooOftenUsedAdjectiveRule(t *testing.T) {
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	sents := languagetool.SplitAndAnalyze("Ein schönes Auto. Noch ein schönes Haus.")
	require.GreaterOrEqual(t, len(rule.MatchList(sents)), 2)
}
