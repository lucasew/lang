package zh

// Twin of ChineseTaggerTest — parses HanLP-encoded surface/pos tokens like Java.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseTagger_Tagger(t *testing.T) {
	// Java path: tokenizer emits "surface/pos"; tagger splits.
	got := NewChineseTagger().Tag([]string{"你好/i", "世界/n", "xyz"})
	require.Len(t, got, 3)
	require.Equal(t, "你好", got[0].GetToken())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "i", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "世界", got[1].GetToken())
	require.Equal(t, "n", *got[1].GetReadings()[0].GetPOSTag())
	// malformed (no slash) → space token like Java
	require.Equal(t, " ", got[2].GetToken())
	require.Equal(t, ChineseDictPath, NewChineseTagger().GetDictionaryPath())
}

// Twin of ChineseTagger.asAnalyzedToken: HanLP unknown tag "x" is kept (not invent nil POS).
func TestChineseTagger_KeepsXPOS(t *testing.T) {
	got := NewChineseTagger().Tag([]string{"未知/x", "词/n"})
	require.Len(t, got, 2)
	require.Equal(t, "未知", got[0].GetToken())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "x", *got[0].GetReadings()[0].GetPOSTag(), "Java keeps POS x")
	require.Equal(t, "n", *got[1].GetReadings()[0].GetPOSTag())
}

// Twin of /w punctuation path: empty surface + trailing w → POS w, surface without /w.
func TestChineseTagger_PunctuationW(t *testing.T) {
	// e.g. "。/w" normal path; special path is surface empty before last /w
	// Java: parts[0]=="" && parts[last]=="w" uses word.substring(0, len-2) and POS "w"
	got := NewChineseTagger().Tag([]string{"/w"})
	require.Len(t, got, 1)
	require.Equal(t, "", got[0].GetToken()) // len("/w")-2 = 0
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "w", *got[0].GetReadings()[0].GetPOSTag())
}
