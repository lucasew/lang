package zh

// Twin of ChineseSentenceTokenizerTest (Java TestTools.testSplit).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testSplit(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewChineseSentenceTokenizer()
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "input=%q", joined)
}

func TestChineseSentenceTokenizer_Tokenize(t *testing.T) {
	t1 := "他说："
	_ = t1
	t2 := "我们是中国人"
	t3 := "中国人很好"

	// punctuation1: do NOT split
	for _, i := range []rune{'_', '/', ':', '@', '#', '$', '%', '^', '&', '+', '*'} {
		testSplit(t, t2+string(i)+t3)
	}
	// punctuation2: split after Chinese clause/sentence marks
	for _, i := range []rune{'，', '！', '？', '；', '。'} {
		testSplit(t, t2+string(i), t3)
	}
}

func TestChineseSentenceTokenizer_Tokenize2(t *testing.T) {
	testSplit(t,
		"Linux是一種自由和開放源碼的類UNIX操作系統。",
		"该操作系统的内核由林纳斯·托瓦兹在1991年10月5日首次发布。",
		"在加上使用者空間的應用程式之後，",
		"成為Linux作業系統。",
	)
}

func TestChineseSentenceTokenizer_TokenizeWithSpaces(t *testing.T) {
	testSplit(t, "的", " ", "诗的。")
	testSplit(t, "的", "  ", "诗的。")
	testSplit(t, "的", "\n", "诗的。")
	testSplit(t, "的", "\n\n", "诗的。")
	testSplit(t, "的", "\n \n", "诗的。")
	testSplit(t, "的", "\n \n")
	testSplit(t, " ", "的", " ")
}

func TestChineseSentenceTokenizer_SingleLineBreaksNoEffect(t *testing.T) {
	tok := NewChineseSentenceTokenizer()
	tok.SetSingleLineBreaksMarksParagraph(true)
	require.False(t, tok.SingleLineBreaksMarksPara())
}
