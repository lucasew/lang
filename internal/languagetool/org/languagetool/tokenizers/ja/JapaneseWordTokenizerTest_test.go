package ja

// Twin of JapaneseWordTokenizerTest — kagome/IPA mirrors Java Sen (lucene-gosen).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJapaneseWordTokenizer_Tokenize(t *testing.T) {
	tok := NewJapaneseWordTokenizer()
	// custom segmenter path
	tok.Segment = func(text string) []string {
		return []string{"日本 名詞-一般 日本", "語 名詞-一般 語"}
	}
	require.Equal(t, []string{"日本 名詞-一般 日本", "語 名詞-一般 語"}, tok.Tokenize("日本語"))

	// Java twin: これはペンです。
	got := NewJapaneseWordTokenizer().Tokenize("これはペンです。")
	require.Equal(t, []string{
		"これ 名詞-代名詞-一般 これ",
		"は 助詞-係助詞 は",
		"ペン 名詞-一般 ペン",
		"です 助動詞 です",
		"。 記号-句点 。",
	}, got)

	// Second Java example (答えた stem 答え + た).
	got2 := NewJapaneseWordTokenizer().Tokenize("私は「うん、そうだ」と答えた。")
	require.Equal(t, []string{
		"私 名詞-代名詞-一般 私",
		"は 助詞-係助詞 は",
		"「 記号-括弧開 「",
		"うん 感動詞 うん",
		"、 記号-読点 、",
		"そう 副詞-助詞類接続 そう",
		"だ 助動詞 だ",
		"」 記号-括弧閉 」",
		"と 助詞-格助詞-引用 と",
		"答え 動詞-自立 答える",
		"た 助動詞 た",
		"。 記号-句点 。",
	}, got2)
}
