package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDynamicAdj_Podibny(t *testing.T) {
	rs := DynamicAdjReadings("Ш-подібному")
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].Lemma, "подібн")
	require.Contains(t, rs[0].POS, "adj")
}

func TestDynamicAdj_Vmisny(t *testing.T) {
	rs := DynamicAdjReadings("карбонат-вмісні")
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "adj")
}

func TestUkrainianTagger_XShaped(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"Ш-подібному", "S-подібної"})
	require.True(t, got[0].IsTagged())
	require.True(t, got[1].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "adj")
}

func TestUkrainianTagger_Vmisny(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"Са-вмісні"})
	require.True(t, got[0].IsTagged())
}

func TestUkrainianTagger_Intj(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"гей-но", "а-а", "гаааа"})
	require.True(t, got[0].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "intj")
	require.True(t, got[1].IsTagged())
	require.True(t, got[2].IsTagged())
}
