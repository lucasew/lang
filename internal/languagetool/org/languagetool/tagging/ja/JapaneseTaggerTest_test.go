package ja

// Twin of org.languagetool.tagging.ja.JapaneseTaggerTest
// (inspiration/languagetool/.../tagging/ja/JapaneseTaggerTest.java).
//
// Java: TestTools.myAssert(input, expected, JapaneseWordTokenizer, JapaneseTagger)
// with exact expected reading strings (sorted per-token readings via Collections.sort).
import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	jatok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ja"
	"github.com/stretchr/testify/require"
)

func TestJapaneseTagger_Tagger(t *testing.T) {
	// Java JapaneseTaggerTest.testTagger — three myAssert cases with exact expected strings.
	cases := []struct {
		input    string
		expected string
	}{
		{
			"これは簡単なテストです。",
			"これ/[これ]名詞-代名詞-一般 -- は/[は]助詞-係助詞 -- 簡単/[簡単]名詞-形容動詞語幹 -- な/[だ]助動詞 -- テスト/[テスト]名詞-サ変接続 -- です/[です]助動詞 -- 。/[。]記号-句点",
		},
		{
			"私は眠い。",
			"私/[私]名詞-代名詞-一般 -- は/[は]助詞-係助詞 -- 眠い/[眠い]形容詞-自立 -- 。/[。]記号-句点",
		},
		{
			"とても冷たい飲み物。",
			"とても/[とても]副詞-助詞類接続 -- 冷たい/[冷たい]形容詞-自立 -- 飲み物/[飲み物]名詞-一般 -- 。/[。]記号-句点",
		},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := myAssertTagger(tc.input)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestJapaneseTagger_AsAnalyzedTokenMalformed(t *testing.T) {
	// Java asAnalyzedToken: parts.length != 3 → AnalyzedToken(" ", null, null)
	tagger := NewJapaneseTagger()
	bad := tagger.Tag([]string{"not-three-parts"})
	require.Len(t, bad, 1)
	require.Equal(t, " ", bad[0].GetToken())
	require.Nil(t, bad[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, bad[0].GetReadings()[0].GetLemma())
	require.Equal(t, 0, bad[0].GetStartPos())
}

func TestJapaneseTagger_StartPosUTF16(t *testing.T) {
	// Java: pos += at.getToken().length() after each token (UTF-16 code units).
	tagger := NewJapaneseTagger()
	got := tagger.Tag([]string{
		"これ 名詞-代名詞-一般 これ",
		"は 助詞-係助詞 は",
	})
	require.Len(t, got, 2)
	require.Equal(t, 0, got[0].GetStartPos())
	// "これ" is two BMP chars → UTF-16 length 2
	require.Equal(t, 2, got[1].GetStartPos())
}

func TestJapaneseTagger_CreateNullAndToken(t *testing.T) {
	tagger := NewJapaneseTagger()
	n := tagger.CreateNullToken("x", 7)
	require.Equal(t, "x", n.GetToken())
	require.Equal(t, 7, n.GetStartPos())
	require.Nil(t, n.GetReadings()[0].GetPOSTag())
	require.Nil(t, n.GetReadings()[0].GetLemma())

	ct := tagger.CreateToken("y", "POS")
	require.Equal(t, "y", ct.GetToken())
	require.NotNil(t, ct.GetPOSTag())
	require.Equal(t, "POS", *ct.GetPOSTag())
	require.Nil(t, ct.GetLemma())
}

// myAssertTagger ports Java TestTools.myAssert(input, expected, tokenizer, tagger):
// tokenize → drop non-word tokens → tag → sorted readings joined by " -- ".
func myAssertTagger(input string) string {
	tokenizer := jatok.NewJapaneseWordTokenizer()
	tagger := NewJapaneseTagger()
	tokens := tokenizer.Tokenize(input)
	var noWS []string
	for _, tok := range tokens {
		if testToolsIsWord(tok) {
			noWS = append(noWS, tok)
		}
	}
	output := tagger.Tag(noWS)
	var parts []string
	for _, atr := range output {
		parts = append(parts, strings.Join(testToolsGetAsStrings(atr), "|"))
	}
	return strings.Join(parts, " -- ")
}

// testToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func testToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// testToolsGetAsStrings ports TestTools.getAsStrings (sorted).
func testToolsGetAsStrings(atr *languagetool.AnalyzedTokenReadings) []string {
	if atr == nil {
		return nil
	}
	var readings []string
	for _, r := range atr.GetReadings() {
		if r != nil {
			readings = append(readings, testToolsGetAsString(r))
		}
	}
	sort.Strings(readings)
	return readings
}

// testToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
func testToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
