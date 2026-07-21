package tokenizers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tokenizers.WordTokenizerTest.

func TestWordTokenizer_Tokenize(t *testing.T) {
	wt := NewWordTokenizer()
	tokens := wt.Tokenize("This is\u00A0a test")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[This,  , is, \u00A0, a,  , test]", tokenListString(tokens))

	tokens = wt.Tokenize("This\rbreaks")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[This, \r, breaks]", tokenListString(tokens))

	tokens = wt.Tokenize("dev.all@languagetool.org")
	require.Equal(t, 1, len(tokens))
	tokens = wt.Tokenize("dev.all@languagetool.org.")
	require.Equal(t, 2, len(tokens))
	tokens = wt.Tokenize("dev.all@languagetool.org:")
	require.Equal(t, 2, len(tokens))
	tokens = wt.Tokenize("Schreiben Sie Hr. Meier (meier@mail.com).")
	require.Equal(t, 13, len(tokens))
	tokens = wt.Tokenize("Get more at languagetool.org/foo, and via twitter")
	require.Equal(t, 14, len(tokens))
	require.Contains(t, tokens, "languagetool.org/foo")
	tokens = wt.Tokenize("Get more at sub.languagetool.org/foo, and via twitter")
	require.Equal(t, 14, len(tokens))
	require.Contains(t, tokens, "sub.languagetool.org/foo")
}

func TestWordTokenizer_IsUrl(t *testing.T) {
	require.True(t, IsURL("www.languagetool.org"))
	require.True(t, IsURL("languagetool.org/"))
	require.True(t, IsURL("languagetool.org/foo"))
	require.True(t, IsURL("subdomain.languagetool.org/"))
	require.True(t, IsURL("http://www.languagetool.org"))
	require.True(t, IsURL("https://www.languagetool.org"))
	require.False(t, IsURL("languagetool.org"))
	require.False(t, IsURL("sub.languagetool.org"))
	require.False(t, IsURL("something-else"))
}

func TestWordTokenizer_IsEMail(t *testing.T) {
	require.True(t, IsEMail("martin.mustermann@test.de"))
	require.True(t, IsEMail("martin.mustermann@test.languagetool.de"))
	require.True(t, IsEMail("martin-mustermann@test.com"))
	require.False(t, IsEMail("@test.de"))
	require.False(t, IsEMail("f.test@test"))
	require.False(t, IsEMail("f@t.t"))
}

func tokenListString(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

func tokenizePipe(s string) string {
	toks := NewWordTokenizer().Tokenize(s)
	return strings.Join(toks, "|")
}

func TestWordTokenizer_UrlTokenize(t *testing.T) {
	require.Equal(t, "\"|This| |http://foo.org|.|\"", tokenizePipe("\"This http://foo.org.\""))
	require.Equal(t, "«|This| |http://foo.org|.|»", tokenizePipe("«This http://foo.org.»"))
	require.Equal(t, "This| |http://foo.org|.|.|.", tokenizePipe("This http://foo.org..."))
	require.Equal(t, "This| |http://foo.org|.", tokenizePipe("This http://foo.org."))
	require.Equal(t, "This| |http://foo.org| |blah", tokenizePipe("This http://foo.org blah"))
	require.Equal(t, "This| |http://foo.org| |and| |ftp://bla.com| |blah", tokenizePipe("This http://foo.org and ftp://bla.com blah"))
	require.Equal(t, "foo| |http://localhost:32000/?ch=1| |bar", tokenizePipe("foo http://localhost:32000/?ch=1 bar"))
	require.Equal(t, "foo| |ftp://localhost:32000/| |bar", tokenizePipe("foo ftp://localhost:32000/ bar"))
	require.Equal(t, "foo| |http://google.de/?aaa| |bar", tokenizePipe("foo http://google.de/?aaa bar"))
	require.Equal(t, "foo| |http://www.flickr.com/123@N04/hallo#test| |bar", tokenizePipe("foo http://www.flickr.com/123@N04/hallo#test bar"))
	require.Equal(t, "foo| |http://www.youtube.com/watch?v=wDN_EYUvUq0| |bar", tokenizePipe("foo http://www.youtube.com/watch?v=wDN_EYUvUq0 bar"))
	require.Equal(t, "foo| |http://example.net/index.html?s=A54C6FE2%23info| |bar", tokenizePipe("foo http://example.net/index.html?s=A54C6FE2%23info bar"))
	require.Equal(t, "foo| |https://writerduet.com/script/#V6922~***~branch=-MClu-LnPrTNz8oz_rJb| |bar",
		tokenizePipe("foo https://writerduet.com/script/#V6922~***~branch=-MClu-LnPrTNz8oz_rJb bar"))
	require.Equal(t, "foo| |https://joe:passwd@example.net:8080/index.html?action=x&session=A54C6FE2#info| |bar",
		tokenizePipe("foo https://joe:passwd@example.net:8080/index.html?action=x&session=A54C6FE2#info bar"))
}

func TestWordTokenizer_UrlTokenizeWithQuote(t *testing.T) {
	require.Equal(t, "This| |'|http://foo.org|'| |blah", tokenizePipe("This 'http://foo.org' blah"))
	require.Equal(t, `This| |"|http://foo.org|"| |blah`, tokenizePipe(`This "http://foo.org" blah`))
	require.Equal(t, `This| |(|"|http://foo.org|"|)| |blah`, tokenizePipe(`This ("http://foo.org") blah`))
}

func TestWordTokenizer_UrlTokenizeWithAppendedCharacter(t *testing.T) {
	require.Equal(t, "foo| |(|http://ex.net/p?a=x#i|)| |bar", tokenizePipe("foo (http://ex.net/p?a=x#i) bar"))
	require.Equal(t, "foo| |http://ex.net/p?a=x#i|,| |bar", tokenizePipe("foo http://ex.net/p?a=x#i, bar"))
	require.Equal(t, "foo| |http://ex.net/p?a=x#i|.| |bar", tokenizePipe("foo http://ex.net/p?a=x#i. bar"))
	require.Equal(t, "foo| |http://ex.net/p?a=x#i|:| |bar", tokenizePipe("foo http://ex.net/p?a=x#i: bar"))
	require.Equal(t, "foo| |http://ex.net/p?a=x#i|?| |bar", tokenizePipe("foo http://ex.net/p?a=x#i? bar"))
	require.Equal(t, "foo| |http://ex.net/p?a=x#i|!| |bar", tokenizePipe("foo http://ex.net/p?a=x#i! bar"))
}

func TestWordTokenizer_IncompleteUrlTokenize(t *testing.T) {
	require.Equal(t, "http|:|/", tokenizePipe("http:/"))
	require.Equal(t, "http://", tokenizePipe("http://"))
	require.Equal(t, "http://a", tokenizePipe("http://a"))
	require.Equal(t, "foo| |http| |bar", tokenizePipe("foo http bar"))
	require.Equal(t, "foo| |http|:| |bar", tokenizePipe("foo http: bar"))
	require.Equal(t, "foo| |http|:|/| |bar", tokenizePipe("foo http:/ bar"))
	require.Equal(t, "foo| |http://| |bar", tokenizePipe("foo http:// bar"))
	require.Equal(t, "foo| |http://a| |bar", tokenizePipe("foo http://a bar"))
	require.Equal(t, "foo| |http://|?| |bar", tokenizePipe("foo http://? bar"))
}

// Twin of WordTokenizerTest.testCheckCurrencyExpression
func TestWordTokenizer_CheckCurrencyExpression(t *testing.T) {
	require.True(t, IsCurrencyExpression("US$45"))
	require.True(t, IsCurrencyExpression("5,000€"))
	require.True(t, IsCurrencyExpression("£1.50"))
	require.True(t, IsCurrencyExpression("R$1.999.99"))
	require.False(t, IsCurrencyExpression("US$"))
	require.False(t, IsCurrencyExpression("X€"))
	require.False(t, IsCurrencyExpression(".50£"))
	require.False(t, IsCurrencyExpression("5R$5"))
}

// Twin of WordTokenizerTest.testSplitCurrencyExpression
func TestWordTokenizer_SplitCurrencyExpression(t *testing.T) {
	require.Equal(t, []string{"US$", "45"}, SplitCurrencyExpression("US$45"))
	require.Equal(t, []string{"5,000", "€"}, SplitCurrencyExpression("5,000€"))
	require.Equal(t, []string{"£", "1.50"}, SplitCurrencyExpression("£1.50"))
	require.Equal(t, []string{"R$", "1.999.99"}, SplitCurrencyExpression("R$1.999.99"))
	// not currency expr — return original token only
	require.Equal(t, []string{"US$X"}, SplitCurrencyExpression("US$X"))
	require.Equal(t, []string{"foobar"}, SplitCurrencyExpression("foobar"))
}
