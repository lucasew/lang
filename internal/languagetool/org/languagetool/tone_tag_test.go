package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRealToneTags(t *testing.T) {
	tags := RealToneTags()
	require.NotContains(t, tags, ToneNoToneRule)
	require.NotContains(t, tags, ToneAllToneRules)
	require.Contains(t, tags, ToneClarity)
	require.Len(t, tags, 13)
}
