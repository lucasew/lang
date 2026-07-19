package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java en.LongSentenceRule.setUrl + getShortMessage; base STYLE/picky.
func TestLongSentenceRule_Metadata(t *testing.T) {
	rule := NewLongSentenceRule(nil, 40)
	require.Equal(t, "TOO_LONG_SENTENCE", rule.GetID())
	require.Equal(t, "Long sentence", rule.ShortMsg)
	require.Contains(t, rule.GetURL(), "splitting-long-sentences")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.True(t, rule.HasTag(rules.TagPicky))
}
