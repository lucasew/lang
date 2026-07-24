package km

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestKhmerUnpairedBracketsRule(t *testing.T) {
	rule := NewKhmerUnpairedBracketsRule(nil)
	require.Equal(t, "KM_UNPAIRED_BRACKETS", rule.GetID())
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("(ok)")})))
	// Opening paren without close is reported when sentence ends like a real sentence.
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("(ok.")})))
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("ok)")})))
}
