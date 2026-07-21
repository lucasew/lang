package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestEkavianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{"raditi": {tagging.NewTaggedWord("raditi", "V")}}
	tagger := NewEkavianTagger(wt)
	got := tagger.TagWord("raditi")
	require.Len(t, got, 1)
	require.Equal(t, "V", got[0].GetPosTag())
}

func TestEkavianTagger_Dictionary(t *testing.T) {
	require.Equal(t, EkavianDictionaryPath, NewEkavianTagger(nil).GetDictionaryPath())
}

// Twin of EkavianTaggerTest.testTaggerRaditi
func TestEkavianTagger_TaggerRaditi(t *testing.T) {
	wt := tagging.MapWordTagger{
		"радим":  {tagging.NewTaggedWord("радити", "GL:GV:PZ:1L:0J")},
		"радећи": {tagging.NewTaggedWord("радити", "PL:PN")},
	}
	tg := NewEkavianTagger(wt)
	got := tg.TagWord("радим")
	require.Len(t, got, 1)
	require.Equal(t, "радити", got[0].GetLemma())
	require.Equal(t, "GL:GV:PZ:1L:0J", got[0].GetPosTag())
	got = tg.TagWord("радећи")
	require.Equal(t, "PL:PN", got[0].GetPosTag())
}

// Twin of EkavianTaggerTest.testTaggerJesam
func TestEkavianTagger_TaggerJesam(t *testing.T) {
	wt := tagging.MapWordTagger{
		"је":   {tagging.NewTaggedWord("јесам", "GL:PM:PZ:3L:0J")},
		"јеси": {tagging.NewTaggedWord("јесам", "GL:PM:PZ:2L:0J")},
		"смо":  {tagging.NewTaggedWord("јесам", "GL:PM:PZ:1L:0M")},
	}
	tg := NewEkavianTagger(wt)
	require.Equal(t, "јесам", tg.TagWord("је")[0].GetLemma())
	require.Equal(t, "GL:PM:PZ:3L:0J", tg.TagWord("је")[0].GetPosTag())
	require.Equal(t, "GL:PM:PZ:2L:0J", tg.TagWord("јеси")[0].GetPosTag())
	require.Equal(t, "GL:PM:PZ:1L:0M", tg.TagWord("смо")[0].GetPosTag())
}
