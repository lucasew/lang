package ja

// Twin of JapaneseTaggerTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestJapaneseTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"猫": {tagging.NewTaggedWord("猫", "名詞")},
		"です": {tagging.NewTaggedWord("です", "助動詞")},
	}
	got := NewJapaneseTagger(wt).Tag([]string{"猫", "です", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
	require.Equal(t, JapaneseDictPath, NewJapaneseTagger(wt).GetDictionaryPath())
}
