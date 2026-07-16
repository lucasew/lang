package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tools.StringToolsTest (pure methods; Language-dependent later).

func TestStringTools_AssureSet(t *testing.T) {
	require.Panics(t, func() { AssureSet("", "varName") })
	require.Panics(t, func() { AssureSet(" \t", "varName") })
	// null N/A in Go for first arg
	AssureSet("foo", "varName")
}

func TestStringTools_IsAllUppercase(t *testing.T) {
	require.True(t, IsAllUppercase("A"))
	require.True(t, IsAllUppercase("ABC"))
	require.True(t, IsAllUppercase("ASV-EDR"))
	require.True(t, IsAllUppercase("ASV-ÖÄÜ"))
	require.True(t, IsAllUppercase(""))
	require.False(t, IsAllUppercase("ß"))
	require.False(t, IsAllUppercase("AAAAAAAAAAAAq"))
	require.False(t, IsAllUppercase("a"))
	require.False(t, IsAllUppercase("abc"))
}

func TestStringTools_IsMixedCase(t *testing.T) {
	require.True(t, IsMixedCase("AbC"))
	require.True(t, IsMixedCase("MixedCase"))
	require.True(t, IsMixedCase("iPod"))
	require.True(t, IsMixedCase("AbCdE"))
	require.False(t, IsMixedCase(""))
	require.False(t, IsMixedCase("ABC"))
	require.False(t, IsMixedCase("abc"))
	require.False(t, IsMixedCase("!"))
	require.False(t, IsMixedCase("Word"))
}

func TestStringTools_IsCapitalizedWord(t *testing.T) {
	require.True(t, IsCapitalizedWord("Abc"))
	require.True(t, IsCapitalizedWord("Uppercase"))
	require.True(t, IsCapitalizedWord("Ipod"))
	require.False(t, IsCapitalizedWord(""))
	require.False(t, IsCapitalizedWord("ABC"))
	require.False(t, IsCapitalizedWord("abc"))
	require.False(t, IsCapitalizedWord("!"))
	require.False(t, IsCapitalizedWord("wOrD"))
}

func TestStringTools_StartsWithUppercase(t *testing.T) {
	require.True(t, StartsWithUppercase("A"))
	require.True(t, StartsWithUppercase("ÄÖ"))
	require.False(t, StartsWithUppercase(""))
	require.False(t, StartsWithUppercase("ß"))
	require.False(t, StartsWithUppercase("-"))
}

func TestStringTools_UppercaseFirstChar(t *testing.T) {
	require.Equal(t, "", UppercaseFirstChar(""))
	require.Equal(t, "A", UppercaseFirstChar("A"))
	require.Equal(t, "Öäü", UppercaseFirstChar("öäü"))
	require.Equal(t, "ßa", UppercaseFirstChar("ßa"))
	require.Equal(t, "'Test'", UppercaseFirstChar("'test'"))
	require.Equal(t, "''Test", UppercaseFirstChar("''test"))
	require.Equal(t, "''T", UppercaseFirstChar("''t"))
	require.Equal(t, "'''", UppercaseFirstChar("'''"))
}

func TestStringTools_LowercaseFirstChar(t *testing.T) {
	require.Equal(t, "", LowercaseFirstChar(""))
	require.Equal(t, "a", LowercaseFirstChar("A"))
	require.Equal(t, "öäü", LowercaseFirstChar("Öäü"))
	require.Equal(t, "ßa", LowercaseFirstChar("ßa"))
	require.Equal(t, "'test'", LowercaseFirstChar("'Test'"))
	require.Equal(t, "''test", LowercaseFirstChar("''Test"))
	require.Equal(t, "''t", LowercaseFirstChar("''T"))
	require.Equal(t, "'''", LowercaseFirstChar("'''"))
}

func TestStringTools_ReaderToString(t *testing.T) {
	str, err := ReaderToString(strings.NewReader("bla\nöäü"))
	require.NoError(t, err)
	require.Equal(t, "bla\nöäü", str)
	longStr := strings.Repeat("x", 4000) + "1234567"
	require.Equal(t, 4007, len(longStr))
	str2, err := ReaderToString(strings.NewReader(longStr))
	require.NoError(t, err)
	require.Equal(t, longStr, str2)
}

func TestStringTools_EscapeXMLandHTML(t *testing.T) {
	require.Equal(t, "foo bar", EscapeXML("foo bar"))
	require.Equal(t, "!ä&quot;&lt;&gt;&amp;&amp;", EscapeXML("!ä\"<>&&"))
	require.Equal(t, "!ä&quot;&lt;&gt;&amp;&amp;", EscapeHTML("!ä\"<>&&"))
}

func TestStringTools_ListToString(t *testing.T) {
	list := []string{"foo", "bar", ","}
	require.Equal(t, "foo,bar,,", strings.Join(list, ","))
	require.Equal(t, "foo\tbar\t,", strings.Join(list, "\t"))
}

func TestStringTools_TrimWhitespace(t *testing.T) {
	require.Equal(t, "", TrimWhitespace(""))
	require.Equal(t, "", TrimWhitespace(" "))
	require.Equal(t, "XXY", TrimWhitespace(" \nXX\t Y"))
	require.Equal(t, "XXY", TrimWhitespace(" \r\nXX\t Y"))
	require.Equal(t, "word", TrimWhitespace("word"))
	require.Equal(t, "1 234,56", TrimWhitespace("1 234,56"))
	require.Equal(t, "1234,56", TrimWhitespace("1  234,56"))
}

func TestStringTools_IsWhitespace(t *testing.T) {
	require.Equal(t, true, IsWhitespace("\uFEFF"))
	require.Equal(t, true, IsWhitespace("  "))
	require.Equal(t, true, IsWhitespace("\t"))
	require.Equal(t, true, IsWhitespace("\u2002"))
	require.Equal(t, true, IsWhitespace("\u00a0"))
	require.Equal(t, false, IsWhitespace("abc"))
	require.Equal(t, false, IsWhitespace("\u0001"))
	require.Equal(t, true, IsWhitespace("\u202F"))
}

func TestStringTools_IsPositiveNumber(t *testing.T) {
	require.Equal(t, true, IsPositiveNumber('3'))
	require.Equal(t, false, IsPositiveNumber('a'))
}

func TestStringTools_IsEmpty(t *testing.T) {
	require.Equal(t, true, IsEmptyStr(""))
	require.Equal(t, false, IsEmptyStr("a"))
}

func TestStringTools_FilterXML(t *testing.T) {
	require.Equal(t, "test", FilterXML("test"))
	require.Equal(t, "<<test>>", FilterXML("<<test>>"))
	require.Equal(t, "test", FilterXML("<b>test</b>"))
	require.Equal(t, "A sentence with a test", FilterXML("A sentence with a <em>test</em>"))
}

func TestStringTools_AllStartWithLowercase(t *testing.T) {
	require.True(t, AllStartWithLowercase("the lord of the rings"))
	require.False(t, AllStartWithLowercase("the Fellowship of the Ring"))
	require.True(t, AllStartWithLowercase("bilbo"))
	require.False(t, AllStartWithLowercase("Baggins"))
}

func TestStringTools_ToId(t *testing.T) {
	// Java String.toUpperCase maps ß→SS (pre-mapped in ToId for Go parity).
	require.Equal(t, "BL_Q_A__UEBEL_OEAESSOE", ToId(" Bl'a (übel öäßÖ ", "de"))
	require.Equal(t, "ÜSS_ÇÃÔ_OÙ_Ñ", ToId("üß çãÔ-où Ñ", "pt"))
	require.Equal(t, "FOOÓÉÉ", ToId("fooóéÉ", "de"))
}

func TestStringTools_ReadStream(t *testing.T) {
	root := findModuleRoot(t)
	path := filepath.Join(root, "inspiration/languagetool/languagetool-core/src/test/resources/testinput.txt")
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	content, err := ReadStream(f)
	require.NoError(t, err)
	require.Equal(t, "one\ntwo\nöäüß\nșțîâăȘȚÎÂĂ\n", content)
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestStringTools_AddSpace(t *testing.T) {
	require.Equal(t, " ", AddSpace("word", "en"))
	require.Equal(t, "", AddSpace(",", "en"))
	require.Equal(t, "", AddSpace(",", "en"))
	require.Equal(t, "", AddSpace(",", "en"))
	require.Equal(t, "", AddSpace(".", "fr"))
	require.Equal(t, "", AddSpace(".", "de"))
	require.Equal(t, " ", AddSpace("!", "fr"))
	require.Equal(t, "", AddSpace("!", "de"))
}

func TestStringTools_AsString(t *testing.T) {
	require.Nil(t, AsString(nil))
	s := "foo!"
	require.Equal(t, &s, AsString(&s))
	require.Equal(t, "foo!", AsStringFromValue("foo!"))
}

func TestStringTools_IsCamelCase(t *testing.T) {
	require.False(t, IsCamelCase("abc"))
	require.False(t, IsCamelCase("ABC"))
	require.True(t, IsCamelCase("iSomething"))
	require.True(t, IsCamelCase("iSomeThing"))
	require.True(t, IsCamelCase("mRNA"))
	require.True(t, IsCamelCase("microRNA"))
	require.True(t, IsCamelCase("microSomething"))
	require.True(t, IsCamelCase("iSomeTHING"))
}

func TestStringTools_StringForSpeller(t *testing.T) {
	arabicChars := "\u064B \u064C \u064D \u064E \u064F \u0650 \u0651 \u0652 \u0670"
	require.Equal(t, arabicChars, StringForSpeller(arabicChars))
	russianChars := "а б в г д е ё ж з и й к л м н о п р с т у ф х ц ч ш щ ъ ы ь э ю я"
	require.Equal(t, russianChars, StringForSpeller(russianChars))
	require.Equal(t, "   Prueva", StringForSpeller("🧡 Prueva"))
	// Java: "\uD83E\uDDE1\uD83D\uDEB4\uD83C\uDFFD♂\uFE0F Prueva"
	emojiStr := "\U0001F9E1\U0001F6B4\U0001F3FD♂\uFE0F Prueva"
	require.Equal(t, "         Prueva", StringForSpeller(emojiStr))
}

func TestStringTools_TitlecaseGlobal(t *testing.T) {
	require.Equal(t, "The Lord of the Rings", TitlecaseGlobal("the lord of the rings"))
	require.Equal(t, "Rhythm and Blues", TitlecaseGlobal("rhythm And blues"))
	require.Equal(t, "Memória de Leitura", TitlecaseGlobal("memória de leitura"))
	require.Equal(t, "Fond du Lac", TitlecaseGlobal("fond du lac"))
	require.Equal(t, "El Niño de las Islas", TitlecaseGlobal("el niño de Las islas"))
}
