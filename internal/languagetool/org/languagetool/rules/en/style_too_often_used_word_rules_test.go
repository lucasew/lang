package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleTooOftenUsedNounRule(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule(nil)
	sents := languagetool.SplitAndAnalyze("The problem was hard. Another problem appeared later.")
	require.GreaterOrEqual(t, len(rule.MatchList(sents)), 2)
}

func TestStyleTooOftenUsedVerbRule(t *testing.T) {
	rule := NewStyleTooOftenUsedVerbRule(nil)
	sents := languagetool.SplitAndAnalyze("They gather quickly. Others gather slowly.")
	require.GreaterOrEqual(t, len(rule.MatchList(sents)), 2)
}

func TestStyleTooOftenUsedAdjectiveRule(t *testing.T) {
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	sents := languagetool.SplitAndAnalyze("A beautiful day. Another beautiful night.")
	require.GreaterOrEqual(t, len(rule.MatchList(sents)), 2)
}
