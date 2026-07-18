package ja

// Twin of JapaneseTaggerTest — parses Sen/kagome-encoded tokens like Java.
import (
	"testing"

	jatok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ja"
	"github.com/stretchr/testify/require"
)

func TestJapaneseTagger_Tagger(t *testing.T) {
	// Java TestTools.myAssert path: tokenize then tag encoded rows.
	tokenizer := jatok.NewJapaneseWordTokenizer()
	tagger := NewJapaneseTagger()

	enc := tokenizer.Tokenize("これは簡単なテストです。")
	got := tagger.Tag(enc)
	require.NotEmpty(t, got)
	// これ / 名詞-代名詞-一般
	require.Equal(t, "これ", got[0].GetToken())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "名詞-代名詞-一般", *got[0].GetReadings()[0].GetPOSTag())

	enc2 := tokenizer.Tokenize("私は眠い。")
	got2 := tagger.Tag(enc2)
	require.Equal(t, "私", got2[0].GetToken())
	require.Equal(t, "名詞-代名詞-一般", *got2[0].GetReadings()[0].GetPOSTag())

	// Malformed row → Java space token
	bad := tagger.Tag([]string{"not-three-parts"})
	require.Equal(t, " ", bad[0].GetToken())

	require.Equal(t, JapaneseDictPath, NewJapaneseTagger().GetDictionaryPath())
}
