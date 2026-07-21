package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestJekavianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{"svijet": {tagging.NewTaggedWord("svijet", "N")}}
	tagger := NewJekavianTagger(wt)
	got := tagger.TagWord("svijet")
	require.Len(t, got, 1)
}

func TestJekavianTagger_Dictionary(t *testing.T) {
	require.Equal(t, JekavianDictionaryPath, NewJekavianTagger(nil).GetDictionaryPath())
}

// Twin of JekavianTaggerTest.testTaggerJesam
func TestJekavianTagger_TaggerJesam(t *testing.T) {
	wt := tagging.MapWordTagger{
		"је":   {tagging.NewTaggedWord("јесам", "GL:PM:PZ:3L:0J")},
		"јеси": {tagging.NewTaggedWord("јесам", "GL:PM:PZ:2L:0J")},
		"смо":  {tagging.NewTaggedWord("јесам", "GL:PM:PZ:1L:0M")},
	}
	tg := NewJekavianTagger(wt)
	require.Equal(t, "јесам", tg.TagWord("је")[0].GetLemma())
	require.Equal(t, "GL:PM:PZ:3L:0J", tg.TagWord("је")[0].GetPosTag())
}

// Twin of JekavianTaggerTest.testTaggerSvijet
func TestJekavianTagger_TaggerSvijet(t *testing.T) {
	wt := tagging.MapWordTagger{
		"свијет": {tagging.NewTaggedWord("свијет", "IM:PA:mu:0J:NO")},
	}
	tg := NewJekavianTagger(wt)
	got := tg.TagWord("свијет")
	require.Len(t, got, 1)
	require.Equal(t, "свијет", got[0].GetLemma())
}
