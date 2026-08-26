package en

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadUSGBVariantMap(t *testing.T) {
	// column 1: British key → American (American speller)
	gbToUS := LoadUSGBVariantMap(1)
	if len(gbToUS) == 0 {
		t.Skip("en/en-US-GB.txt not discoverable")
	}
	// colour is GB; color is US — column 1 keys are GB forms
	require.Equal(t, "color", gbToUS["colour"])
	// column 0: American key → British
	usToGB := LoadUSGBVariantMap(0)
	require.Equal(t, "colour", usToGB["color"])

	// Official en-US-GB.txt has commented pairs e.g. "#psychoanalyzes;psychoanalyses".
	// Java loadWordlist skips lines starting with '#' after trim — must not pollute maps.
	// Without the guard, column 1 would insert psychoanalyses → #psychoanalyzes.
	_, hasPsychoanalysesGB := gbToUS["psychoanalyses"]
	require.False(t, hasPsychoanalysesGB, "commented pair must not load GB key psychoanalyses")
	_, hasPsychoanalyzesHash := usToGB["#psychoanalyzes"]
	require.False(t, hasPsychoanalyzesHash, "commented pair must not load key #psychoanalyzes")
	for k, v := range gbToUS {
		require.False(t, strings.HasPrefix(k, "#"), "GB→US key must not start with #: %q", k)
		require.False(t, strings.HasPrefix(v, "#"), "GB→US value must not start with #: %q→%q", k, v)
	}
	for k, v := range usToGB {
		require.False(t, strings.HasPrefix(k, "#"), "US→GB key must not start with #: %q", k)
		require.False(t, strings.HasPrefix(v, "#"), "US→GB value must not start with #: %q→%q", k, v)
	}
}

func TestAmericanSpeller_ColourVariant(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	if len(r.OtherVariant) == 0 {
		t.Skip("en-US-GB not loaded")
	}
	vi := r.IsValidInOtherVariant("colour")
	require.NotNil(t, vi)
	require.Equal(t, "British English", vi.GetVariantName())
	require.Equal(t, "color", vi.GetOtherVariant())
}
