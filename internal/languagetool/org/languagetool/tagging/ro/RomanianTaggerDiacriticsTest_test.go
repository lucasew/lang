package ro

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestRomanianTaggerDiacritics_TaggerMerseseram(t *testing.T) {
	wt := tagging.MapWordTagger{"merseserăm": {tagging.NewTaggedWord("merge", "V")}}
	got := NewRomanianTagger(wt).Tag([]string{"merseserăm"})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestRomanianTaggerDiacritics_TaggerCuscaCutit(t *testing.T) {
	wt := tagging.MapWordTagger{
		"cușcă": {tagging.NewTaggedWord("cușcă", "S")},
		"cuțit": {tagging.NewTaggedWord("cuțit", "S")},
	}
	got := NewRomanianTagger(wt).Tag([]string{"cușcă", "cuțit"})
	require.Len(t, got, 2)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
}
