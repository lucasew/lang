package uk

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDynamicAdj_Podibny(t *testing.T) {
	// Fail-closed without dict
	require.Nil(t, DynamicAdjReadings("Ш-подібному", nil))

	tagWord := func(s string) []tagging.TaggedWord {
		if strings.EqualFold(s, "подібному") {
			return []tagging.TaggedWord{
				{Lemma: "подібний", PosTag: "adj:m:v_dav"},
				{Lemma: "подібний", PosTag: "adj:m:v_mis"},
				{Lemma: "подібний", PosTag: "adj:n:v_dav"},
				{Lemma: "подібний", PosTag: "adj:n:v_mis"},
			}
		}
		return nil
	}
	rs := DynamicAdjReadings("Ш-подібному", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "Ш-подібний", rs[0].Lemma)
	require.Contains(t, rs[0].POS, "adj")
}

func TestDynamicAdj_Vmisny(t *testing.T) {
	// Java: tag "боро"+right, lemma "вмісний"
	tagWord := func(s string) []tagging.TaggedWord {
		if strings.HasPrefix(strings.ToLower(s), "боровмісн") {
			return []tagging.TaggedWord{{Lemma: "боровмісний", PosTag: "adj:p:v_naz"}}
		}
		return nil
	}
	rs := DynamicAdjReadings("карбонат-вмісні", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "карбонат-вмісний", rs[0].Lemma)
	require.Contains(t, rs[0].POS, "adj")
	// without borо* hits: fail closed
	require.Nil(t, DynamicAdjReadings("карбонат-вмісні", func(string) []tagging.TaggedWord { return nil }))
}

func TestUkrainianTagger_XShaped(t *testing.T) {
	wt := tagging.MapWordTagger{
		"подібному": {tagging.NewTaggedWord("подібний", "adj:m:v_dav")},
		"подібної":  {tagging.NewTaggedWord("подібний", "adj:f:v_rod")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.Tag([]string{"Ш-подібному", "S-подібної"})
	require.True(t, got[0].IsTagged())
	require.True(t, got[1].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "adj")

	// empty dict: fail closed
	bare := NewUkrainianTagger(tagging.MapWordTagger{}).Tag([]string{"Ш-подібному"})
	require.False(t, bare[0].IsTagged())
}

func TestUkrainianTagger_Vmisny(t *testing.T) {
	wt := tagging.MapWordTagger{
		"боровмісні": {tagging.NewTaggedWord("боровмісний", "adj:p:v_naz")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.Tag([]string{"Са-вмісні"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "Са-вмісний", *got[0].GetReadings()[0].GetLemma())
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
