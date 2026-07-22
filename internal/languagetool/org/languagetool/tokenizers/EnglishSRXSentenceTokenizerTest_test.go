package tokenizers

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/tokenizers/EnglishSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer: single line breaks mark paragraph = true
// stokenizer2: single line breaks mark paragraph = false
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitEN mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer)
// with stokenizer.setSingleLineBreaksMarksParagraph(true).
func testSplitEN(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewEnglishSRXSentenceTokenizer()
	tok.SetSingleLineBreaksMarksParagraph(true)
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// testSplitEN2 mirrors TestTools.testSplit(sentences, stokenizer2)
// with stokenizer2.setSingleLineBreaksMarksParagraph(false).
func testSplitEN2(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewEnglishSRXSentenceTokenizer()
	tok.SetSingleLineBreaksMarksParagraph(false)
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

func TestEnglishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// incomplete sentences, need to work for on-the-fly checking of texts:
	testSplitEN(t, "What is the I.S?")
	testSplitEN(t, "Where are the I.S and the M.Z notes? ")
	testSplitEN(t, "Here's a")
	testSplitEN(t, "Here's a sentence. ", "And here's one that's not comp")
	testSplitEN(t, "Or did you install it (i.e. MS Word) yourself?")

	testSplitEN(t, "This is a sentence. ")
	testSplitEN(t, "This is a sentence. ", "And this is another one.")
	testSplitEN(t, "This is it. ", "and this is another sentence.")
	testSplitEN(t, "This is a sentence. ", "and this is another sentence.")
	testSplitEN(t, "This is a sentence.", "Isn't it?", "Yes, it is.")
	testSplitEN(t, "This is e.g. Mr. Smith, who talks slowly...",
		"But this is another sentence.")
	testSplitEN(t, "Chanel no. 5 is blah.")
	testSplitEN(t, "Mrs. Jones gave Peter $4.5, to buy Chanel No 5.",
		"He never came back.")
	testSplitEN(t, "On p. 6 there's nothing. ", "Another sentence.")
	testSplitEN(t, "Leave me alone!, he yelled. ", "Another sentence.")
	testSplitEN(t, "\"Leave me alone!\", he yelled.")
	testSplitEN(t, "'Leave me alone!', he yelled. ", "Another sentence.")
	testSplitEN(t, "'Leave me alone!,' he yelled. ", "Another sentence.")
	testSplitEN(t, "This works on the phrase level, i.e. not on the word level.")
	testSplitEN(t, "Let's meet at 5 p.m. in the main street.")
	testSplitEN(t, "James comes from the U.K. where he worked as a programmer.")
	testSplitEN(t, "Don't split strings like U.S.A. please.")
	testSplitEN(t, "Hello ( Hi! ) my name is Chris.")
	testSplitEN(t, "Don't split strings like U. S. A. either.")
	testSplitEN(t, "Don't split strings like U.S.A either.")
	testSplitEN(t, "Don't split... ", "Well you know. ", "Here comes more text.")
	testSplitEN(t, "Don't split... well you know. ", "Here comes more text.")
	testSplitEN(t, "The \".\" should not be a delimiter in quotes.")
	testSplitEN(t, "\"Here he comes!\" she said.")
	testSplitEN(t, "\"Here he comes!\", she said.")
	testSplitEN(t, "\"Here he comes.\" ", "But this is another sentence.")
	testSplitEN(t, "\"Here he comes!\". ", "That's what he said.")
	testSplitEN(t, "The sentence ends here. ", "(Another sentence.)")
	testSplitEN(t, "The sentence (...) ends here.")
	testSplitEN(t, "The sentence [...] ends here.")
	testSplitEN(t, "The sentence ends here (...). ", "Another sentence.")
	// previously known failed but not now :)
	testSplitEN(t, "He won't. ", "Really.")
	testSplitEN(t, "He will not. ", "Really.")
	testSplitEN(t, "He won't go. ", "Really.")
	testSplitEN(t, "He won't say no.", "Not really.")
	testSplitEN(t, "He won't say No.", "Not really.")
	testSplitEN(t, "He won't say no. 5 is better. ", "Not really.")
	testSplitEN(t, "He won't say No. 5 is better. ", "Not really.")
	testSplitEN(t, "They met at 5 p.m. on Thursday.")
	testSplitEN(t, "They met at 5 p.m. ", "It was Thursday.")
	testSplitEN(t, "This is it: a test.")
	testSplitEN(t, "12) Make sure that the lamp is on. ", "12) Make sure that the lamp is on. ")
	testSplitEN(t, "He also offers a conversion table (see Cohen, 1988, p. 123). ")
	// one/two returns = paragraph = new sentence:
	// TestTools.testSplit(..., stokenizer2) — single line breaks do NOT mark paragraph
	testSplitEN2(t, "He won't\n\n", "Really.")
	// TestTools.testSplit(..., stokenizer) — single line breaks mark paragraph
	testSplitEN(t, "He won't\n", "Really.")
	testSplitEN2(t, "He won't\n\n", "Really.")
	testSplitEN2(t, "He won't\nReally.")
	// Missing space after sentence end:
	testSplitEN(t, "James is from the Ireland!", "He lives in Spain now.")
	// From the abbreviation list:
	testSplitEN(t, "Jones Bros. have built a successful company.")
	// parentheses:
	testSplitEN(t, "It (really!) works.")
	testSplitEN(t, "It [really!] works.")
	testSplitEN(t, "It works (really!). ", "No doubt.")
	testSplitEN(t, "It works [really!]. ", "No doubt.")
	testSplitEN(t, "It really(!) works well.")
	testSplitEN(t, "It really[!] works well.")
	// try to deal with at least some nbsp that appear in strange places (e.g. Google Docs, web editors)
	testSplitEN(t, "A test.\u00A0\n", "Another test.")
	// not clear whether this is the best behavior...
	testSplitEN(t, "A test.\u00A0", "Another test.")
	testSplitEN(t, "A test.\n", "Another test.")
	testSplitEN(t, "A test. \n", "Another test.")
	testSplitEN(t, "A test. \n", "\n", "Another test.")
	testSplitEN(t, "\"Here he comes.\"\u00a0", "But this is another sentence.")

	testSplitEN(t, "The new Yahoo! product is nice.")
	testSplitEN(t, "Yahoo!, what is it?")
	testSplitEN(t, "Yahoo!", "What is it?")

	// footnotes in LibOO/OOo look like this
	testSplitEN(t, "This is a sentence.\u0002 ", "And this is another one.")

	testSplitEN(t, "Other good editions are in vol. 4.")
	testSplitEN(t, "Other good editions are in vol. IX.")
	testSplitEN(t, "Other good editions are in vol. I think.") // ambiguous
	testSplitEN(t, "Who Shall I Say is Calling & Other Stories S. Deziemianowicz, ed. (2009)")
	testSplitEN(t, "Who Shall I Say is Calling & Other Stories S. Deziemianowicz, ed. ", "And this is another one.")
	testSplitEN(t, "This is a sentence written by Ed. ", "And this is another one.")
}
