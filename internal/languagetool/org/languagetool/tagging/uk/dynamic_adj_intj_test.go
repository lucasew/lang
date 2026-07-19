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
	// Java: hyphen intj via compound dict tags; elongated via collapse + dict + :alt.
	// No soft invent list (гей-но / а-а without dict stay untagged).
	wt := tagging.MapWordTagger{
		"а":  {tagging.NewTaggedWord("а", "intj")},
		"га": {tagging.NewTaggedWord("га", "intj")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.Tag([]string{"гей-но", "а-а", "гаааа"})
	require.False(t, got[0].IsTagged(), "гей-но needs dict/compound list — fail closed")
	require.True(t, got[1].IsTagged())
	require.Contains(t, *got[1].GetReadings()[0].GetPOSTag(), "intj")
	require.True(t, got[2].IsTagged())
	require.Contains(t, *got[2].GetReadings()[0].GetPOSTag(), "alt")
}

func TestElongatedAltReadings(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "га" {
			return []tagging.TaggedWord{tagging.NewTaggedWord("га", "intj")}
		}
		return nil
	}
	rs := ElongatedAltReadings("гаааа", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "intj")
	require.Contains(t, *rs[0].GetPOSTag(), "alt")
	require.Empty(t, ElongatedAltReadings("гаааа", nil))
	require.Empty(t, ElongatedAltReadings("ііі", tag))
}
