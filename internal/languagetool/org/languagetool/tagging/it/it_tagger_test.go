package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestItalianTagger(t *testing.T) {
	wt := tagging.MapWordTagger{"cane": {tagging.NewTaggedWord("cane", "S")}}
	got := NewItalianTagger(wt).Tag([]string{"cane"})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}
