package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestCompoundTagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"будинок": {tagging.NewTaggedWord("будинок", "noun:inanim:m:v_naz")},
	}
	inner := NewUkrainianTagger(wt)
	dbg := NewCompoundDebugLogger(true)
	ct := NewCompoundTagger(inner)
	ct.Debug = dbg
	got := ct.Tag([]string{"міні-будинок"})
	require.Len(t, got, 1)
	require.NotEmpty(t, dbg.Lines)
	require.True(t, got[0].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "noun")
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "comp")
}

func TestCompoundTagger_NumericPrefix(t *testing.T) {
	wt := tagging.MapWordTagger{
		"річний": {tagging.NewTaggedWord("річний", "adj:m:v_naz")},
	}
	ct := NewCompoundTagger(NewUkrainianTagger(wt))
	got := ct.Tag([]string{"2-річний"})
	require.True(t, got[0].IsTagged())
}
