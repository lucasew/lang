package zh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Probes match HanLP portable-1.8.2 SentencesUtil (JVM verified).
func TestSentencesUtil_ToSentenceList(t *testing.T) {
	require.Equal(t, []string{"我们是中国人，", "中国人很好"}, sentencesUtilToSentenceList("我们是中国人，中国人很好"))
	require.Equal(t, []string{"Hello.", "世界"}, sentencesUtilToSentenceList("Hello.世界"))
	require.Equal(t, []string{"Hello.World"}, sentencesUtilToSentenceList("Hello.World"))
	require.Equal(t, []string{"Hello."}, sentencesUtilToSentenceList("Hello."))
	require.Equal(t, []string{"Hi!", "There"}, sentencesUtilToSentenceList("Hi!There"))
	require.Equal(t, []string{"Hi?", "There"}, sentencesUtilToSentenceList("Hi?There"))
	require.Equal(t, []string{"等等……", "然后"}, sentencesUtilToSentenceList("等等……然后"))
	require.Equal(t, []string{"等等…然后"}, sentencesUtilToSentenceList("等等…然后"))
	require.Equal(t, []string{"a,", "b"}, sentencesUtilToSentenceList("a,b"))
	require.Equal(t, []string{"a;", "b"}, sentencesUtilToSentenceList("a;b"))
	require.Equal(t, []string{"a，", "b"}, sentencesUtilToSentenceList("a，b"))
	require.Equal(t, []string{"a；", "b"}, sentencesUtilToSentenceList("a；b"))
	require.Equal(t, []string{"a\u00A0", "b"}, sentencesUtilToSentenceList("a\u00A0b"))
	// tab reaches util: appended then split; trim strips tab → "a","b"
	require.Equal(t, []string{"a", "b"}, sentencesUtilToSentenceList("a\tb"))
	require.Equal(t, []string{"我们是中国人_中国人很好"}, sentencesUtilToSentenceList("我们是中国人_中国人很好"))
	require.Equal(t, []string{"foo。", "bar"}, sentencesUtilToSentenceList("  foo。  bar  "))
	require.Nil(t, sentencesUtilToSentenceList("   "))
	require.Equal(t, []string{"3.14是pi"}, sentencesUtilToSentenceList("3.14是pi"))
	require.Equal(t, []string{"end.", "下"}, sentencesUtilToSentenceList("end.下"))

	require.Equal(t, []string{"a,b;c，d；e"}, sentencesUtilToSentenceListInsert("a,b;c，d；e", false))
	require.Equal(t, []string{"a,", "b;", "c，", "d；", "e"}, sentencesUtilToSentenceListInsert("a,b;c，d；e", true))
}

func TestSentencesUtil_LinuxParagraph(t *testing.T) {
	text := "Linux是一種自由和開放源碼的類UNIX操作系統。" +
		"该操作系统的内核由林纳斯·托瓦兹在1991年10月5日首次发布。" +
		"在加上使用者空間的應用程式之後，" +
		"成為Linux作業系統。"
	want := []string{
		"Linux是一種自由和開放源碼的類UNIX操作系統。",
		"该操作系统的内核由林纳斯·托瓦兹在1991年10月5日首次发布。",
		"在加上使用者空間的應用程式之後，",
		"成為Linux作業系統。",
	}
	require.Equal(t, want, sentencesUtilToSentenceList(text))
}

func TestChineseSentenceTokenizer_EdgeCases_JVM(t *testing.T) {
	st := NewChineseSentenceTokenizer()
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"  \n", []string{"  \n"}},
		{"！", []string{"！"}},
		{"a!", []string{"a!"}},
		{"a.", []string{"a."}},
		{"a.b", []string{"a.b"}},
		{"a.中", []string{"a.", "中"}},
		{"……", []string{"……"}},
		{"…", []string{"…"}},
		{"x……y", []string{"x……", "y"}},
		{"说：我们。", []string{"说：我们。"}},
		{"，，", []string{"，", "，"}},
		{"a\u00A0\u00A0b", []string{"a\u00A0", "\u00A0", "b"}},
		{"Hello! World?", []string{"Hello!", " ", "World?"}},
	}
	for _, tc := range cases {
		got := st.Tokenize(tc.in)
		require.Equal(t, tc.want, got, "input=%q", tc.in)
	}
}
