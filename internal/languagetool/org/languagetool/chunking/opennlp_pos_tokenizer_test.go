package chunking

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestLoadGISModel_POS(t *testing.T) {
	p := DiscoverOpenNLPPOSModel()
	if p == "" {
		t.Skip("en-pos-maxent.bin not found")
	}
	m, err := LoadGISModelFromZip(p)
	require.NoError(t, err)
	require.Greater(t, m.NumOutcomes(), 20)
	probs := m.Eval([]string{"default", "w=the", "suf=e", "pre=t"})
	var sum float64
	for _, p := range probs {
		sum += p
	}
	require.InDelta(t, 1.0, sum, 1e-6)
}

func TestPOSTaggerME_Basic(t *testing.T) {
	p := DiscoverOpenNLPPOSModel()
	if p == "" {
		t.Skip("en-pos-maxent.bin not found")
	}
	tagger, err := NewPOSTaggerME(p)
	require.NoError(t, err)
	tags := tagger.Tag([]string{"The", "quick", "brown", "fox", "jumps"})
	require.Len(t, tags, 5)
	require.Equal(t, "DT", tags[0])
	// fox should be NN-ish; jumps VBZ-ish
	require.True(t, tags[3] == "NN" || tags[3] == "NNP", "fox=%s", tags[3])
	require.True(t, strings.HasPrefix(tags[4], "VB") || tags[4] == "NNS", "jumps=%s", tags[4])
}

func TestTokenizerME_Basic(t *testing.T) {
	p := DiscoverOpenNLPTokenModel()
	if p == "" {
		t.Skip("en-token.bin not found")
	}
	tok, err := NewTokenizerME(p)
	require.NoError(t, err)
	// alphanumeric opt keeps plain words whole
	got := tok.Tokenize("Hello world")
	require.Equal(t, []string{"Hello", "world"}, got)
	// punctuation split
	got = tok.Tokenize("Hello, world.")
	require.Contains(t, got, "Hello")
	require.Contains(t, got, ",")
	require.Contains(t, got, "world")
	require.Contains(t, got, ".")
}

func TestEnglishChunker_OpenNLPFull_JavaParitySample(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" || DiscoverOpenNLPPOSModel() == "" || DiscoverOpenNLPTokenModel() == "" {
		t.Skip("OpenNLP models missing")
	}
	// Java EnglishChunkerTest.testAddChunkTags with createReadingsList-style tokens
	// (word + space alternating, ending without space).
	sentence := "A short test of the bicycle is needed"
	tokens := readingsFromSpacedWords(sentence)
	NewEnglishChunker().AddChunkTags(tokens)

	// "A short test" → B/I/E-NP-singular (after EnglishChunkFilter)
	// indices: 0=A 1=sp 2=short 3=sp 4=test ...
	require.Equal(t, []string{"B-NP-singular"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"I-NP-singular"}, tokens[2].GetChunkTags())
	require.Equal(t, []string{"E-NP-singular"}, tokens[4].GetChunkTags())
	// "the bicycle"
	require.Equal(t, []string{"B-NP-singular"}, tokens[8].GetChunkTags())
	require.Equal(t, []string{"E-NP-singular"}, tokens[10].GetChunkTags())
	// "is needed"
	require.Equal(t, []string{"B-VP"}, tokens[12].GetChunkTags())
	require.Equal(t, []string{"I-VP"}, tokens[14].GetChunkTags())
}

// readingsFromSpacedWords builds ATR list like Java createReadingsList: words and spaces.
func readingsFromSpacedWords(sentence string) []*languagetool.AnalyzedTokenReadings {
	parts := strings.Fields(sentence)
	var out []*languagetool.AnalyzedTokenReadings
	pos := 0
	for i, w := range parts {
		nn := "NN"
		out = append(out, languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken(w, &nn, nil), pos))
		pos += len(w)
		if i < len(parts)-1 {
			sp := " "
			out = append(out, languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(" ", &sp, nil), pos))
			pos++
		}
	}
	return out
}

func TestOpenNLPIsWhitespace_JavaStringUtil(t *testing.T) {
	// opennlp.tools.util.StringUtil.isWhitespace:
	// Character.isWhitespace || SPACE_SEPARATOR (Zs)
	require.True(t, openNLPIsWhitespace(' '))
	require.True(t, openNLPIsWhitespace('\t'))
	require.True(t, openNLPIsWhitespace('\n'))
	// NBSP is Zs — included (OpenNLP docs: no-break spaces are whitespace)
	require.True(t, openNLPIsWhitespace('\u00A0'))
	require.True(t, openNLPIsWhitespace('\u2007')) // figure space Zs
	require.True(t, openNLPIsWhitespace('\u202F')) // narrow NBSP Zs
	require.False(t, openNLPIsWhitespace('a'))
	require.False(t, openNLPIsWhitespace('.'))
}
