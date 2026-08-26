package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.ToneTag — REAL_TONE_TAGS excludes meta tags.

func TestRealToneTags(t *testing.T) {
	// Full enum order from ToneTag.java
	all := []ToneTag{
		ToneClarity, ToneFormal, ToneProfessional, ToneConfident, ToneAcademic,
		TonePovRem, ToneScientific, ToneObjective, TonePersuasive, ToneInformal,
		TonePovAdd, TonePositive, ToneGeneral,
		ToneNoToneRule, ToneAllToneRules, ToneAllWithoutGoalSpecific,
	}
	require.Len(t, all, 16)

	tags := RealToneTags()
	require.Len(t, tags, 13)
	require.NotContains(t, tags, ToneNoToneRule)
	require.NotContains(t, tags, ToneAllToneRules)
	require.NotContains(t, tags, ToneAllWithoutGoalSpecific)
	require.Equal(t, []ToneTag{
		ToneClarity, ToneFormal, ToneProfessional, ToneConfident, ToneAcademic,
		TonePovRem, ToneScientific, ToneObjective, TonePersuasive, ToneInformal,
		TonePovAdd, TonePositive, ToneGeneral,
	}, tags)
	// string values match Java enum names (lowercase for real tones)
	require.Equal(t, "clarity", string(ToneClarity))
	require.Equal(t, "NO_TONE_RULE", string(ToneNoToneRule))
	require.Equal(t, "ALL_TONE_RULES", string(ToneAllToneRules))
	require.Equal(t, "ALL_WITHOUT_GOAL_SPECIFIC", string(ToneAllWithoutGoalSpecific))
}
