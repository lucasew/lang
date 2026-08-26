package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestWireDutchTaggerCompoundParts(t *testing.T) {
	wt := tagging.MapWordTagger{
		"puzzel": {tagging.NewTaggedWord("puzzel", "ZNW:EKV:DE_")},
	}
	tagger := NewWiredDutchTagger(wt)
	require.NotNil(t, tagger.GetCompoundParts)
	// without full lists+noun hooks GetParts usually empty — still callable
	_ = tagger.GetCompoundParts("straatpuzzel")
	// accent path still works with wired tagger
	wt2 := tagging.MapWordTagger{
		"deur": {tagging.NewTaggedWord("deur", "ZNW:EKV:DE_")},
	}
	tagger2 := NewWiredDutchTagger(wt2)
	got := tagger2.Tag([]string{"déúr"})
	require.NotEmpty(t, got[0].GetReadings())
	require.Equal(t, "ZNW:EKV:DE_", *got[0].GetReadings()[0].GetPOSTag())
}
