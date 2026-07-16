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
}
