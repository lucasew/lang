package cache

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestCacheUtilsRoundTrip(t *testing.T) {
	sent := languagetool.AnalyzePlain("colour")
	m := rules.NewRuleMatch(rules.NewFakeRule("MORFO"), sent, 0, 6, "spelling")
	m.SetSuggestedReplacements([]string{"color"})
	c := SerializeResultMatch(m)
	require.Equal(t, "MORFO", c.Rule.ID)
	require.Equal(t, 0, c.OffsetPosition.Start)
	require.Equal(t, 6, c.OffsetPosition.End)

	back := DeserializeResultMatch(c, sent)
	require.Equal(t, "spelling", back.Message)
	require.Equal(t, []string{"color"}, back.GetSuggestedReplacements())

	raw, err := MarshalResultMatchJSON(m)
	require.NoError(t, err)
	m2, err := UnmarshalResultMatchJSON(raw, sent)
	require.NoError(t, err)
	require.Equal(t, m.FromPos, m2.FromPos)
}
