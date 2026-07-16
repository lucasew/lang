package zh

// Twin of ChineseTaggerTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestChineseTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"你好": {tagging.NewTaggedWord("你好", "i")},
		"世界": {tagging.NewTaggedWord("世界", "n")},
	}
	got := NewChineseTagger(wt).Tag([]string{"你好", "世界", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
