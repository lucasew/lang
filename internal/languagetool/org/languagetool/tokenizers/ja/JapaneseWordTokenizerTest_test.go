package ja

// Twin of org.languagetool.tokenizers.ja.JapaneseWordTokenizerTest
// (inspiration/languagetool/.../tokenizers/ja/JapaneseWordTokenizerTest.java).
// Asserts full Java-visible token lists (surface POS basicForm), not smoke-only.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJapaneseWordTokenizer_Tokenize(t *testing.T) {
	w := NewJapaneseWordTokenizer()

	// Java: tokenize("これはペンです。") → size 5, exact list toString
	testList := w.Tokenize("これはペンです。")
	require.Equal(t, 5, len(testList))
	require.Equal(t, []string{
		"これ 名詞-代名詞-一般 これ",
		"は 助詞-係助詞 は",
		"ペン 名詞-一般 ペン",
		"です 助動詞 です",
		"。 記号-句点 。",
	}, testList)

	// Java: tokenize("私は「うん、そうだ」と答えた。") → size 12, exact list
	testList = w.Tokenize("私は「うん、そうだ」と答えた。")
	require.Equal(t, 12, len(testList))
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
	}, testList)
}
