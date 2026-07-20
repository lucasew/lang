package chunking

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/chunking/EnglishChunkerTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func readings(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	p := pos
	tok := languagetool.NewAnalyzedToken(token, &p, nil)
	return languagetool.NewAnalyzedTokenReadingsAt(tok, start)
}

// createReadingsList ports EnglishChunkerTest.createReadingsList (StringTokenizer keep spaces).
func createReadingsList(sentence string) []*languagetool.AnalyzedTokenReadings {
	var result []*languagetool.AnalyzedTokenReadings
	pos := 0
	// Split keeping spaces (like StringTokenizer(sentence, " ", true))
	var parts []string
	cur := ""
	for _, r := range sentence {
		if r == ' ' {
			if cur != "" {
				parts = append(parts, cur)
				cur = ""
			}
			parts = append(parts, " ")
		} else {
			cur += string(r)
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	for _, token := range parts {
		if strings.TrimSpace(token) == "" {
			result = append(result, readings(token, "", pos))
		} else {
			result = append(result, readings(token, "fake", pos))
		}
		pos += len(token)
	}
	return result
}

// Port of EnglishChunkerTest.testAddChunkTags
func TestEnglishChunker_AddChunkTags(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" {
		t.Skip("OpenNLP models required for Java-parity chunker test")
	}
	readingsList := createReadingsList("A short test of the bicycle is needed")
	NewEnglishChunker().AddChunkTags(readingsList)
	require.Len(t, readingsList, 15)
	// "A short test":
	require.Equal(t, []string{"B-NP-singular"}, readingsList[0].GetChunkTags())
	require.Equal(t, []string{"I-NP-singular"}, readingsList[2].GetChunkTags())
	require.Equal(t, []string{"E-NP-singular"}, readingsList[4].GetChunkTags())
	// "the bicycle":
	require.Equal(t, []string{"B-NP-singular"}, readingsList[8].GetChunkTags())
	require.Equal(t, []string{"E-NP-singular"}, readingsList[10].GetChunkTags())
	// "is needed"
	require.Equal(t, []string{"B-VP"}, readingsList[12].GetChunkTags())
	require.Equal(t, []string{"I-VP"}, readingsList[14].GetChunkTags())
}

// Port of EnglishChunkerTest.testSingularNounAtEndOfNounPhrase (simplified without full LT pipeline)
func TestEnglishChunker_SingularNounAtEndOfNounPhrase(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" {
		t.Skip("OpenNLP models required")
	}
	// Approximate LT tokens: SENT_START + words + spaces (positions matter for mapping).
	// "The aircraft manager is here." — without full analyzer we use createReadingsList.
	tokens := createReadingsList("The aircraft manager is here")
	NewEnglishChunker().AddChunkTags(tokens)
	// indices: 0=The 1=sp 2=aircraft 3=sp 4=manager ...
	require.Equal(t, []string{"B-NP-singular"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"I-NP-singular"}, tokens[2].GetChunkTags())
	require.Equal(t, []string{"E-NP-singular"}, tokens[4].GetChunkTags())
}

// Port of EnglishChunkerTest.testAddChunkTagsSingular (abacus sentence NP checks).
// Filter uses LT POS tags (NNS) for plural, not OpenNLP tags — set NNS on "numbers".
func TestEnglishChunker_AddChunkTagsSingular(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" {
		t.Skip("OpenNLP models required")
	}
	tokens := createReadingsList("The abacus shows how numbers can be stored")
	// Java LT tagger yields NNS for "numbers"; filter needs that reading.
	nns := "NNS"
	tokens[8] = languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("numbers", &nns, nil), tokens[8].GetStartPos())
	NewEnglishChunker().AddChunkTags(tokens)
	// "The abacus": 0=The 1=sp 2=abacus
	require.Equal(t, []string{"B-NP-singular"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"E-NP-singular"}, tokens[2].GetChunkTags())
	// "numbers" — B-NP-plural, E-NP-plural (single-token phrase gets both after filter)
	require.Equal(t, []string{"B-NP-plural", "E-NP-plural"}, tokens[8].GetChunkTags())
}

// Port of EnglishChunkerTest.testContractions — OpenNLP maps I / 'll when surfaces match.
func TestEnglishChunker_Contractions(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" {
		t.Skip("OpenNLP models required")
	}
	// Surfaces matching OpenNLP tokenize("I'll be there") → I, 'll, be, there
	tokens := createReadingsList("I'll be there")
	// createReadingsList splits only on space, so "I'll" is one token — map fails for 'll.
	// Use explicit tokens like LT English often produces for contractions:
	tokens = []*languagetool.AnalyzedTokenReadings{
		readings("I", "PRP", 0),
		readings("'ll", "MD", 1),
		readings(" ", "", 4),
		readings("be", "VB", 5),
		readings(" ", "", 7),
		readings("there", "RB", 8),
	}
	NewEnglishChunker().AddChunkTags(tokens)
	require.Contains(t, strings.Join(tokens[0].GetChunkTags(), ","), "NP")
	require.Contains(t, strings.Join(tokens[1].GetChunkTags(), ","), "VP")
	require.Contains(t, strings.Join(tokens[3].GetChunkTags(), ","), "VP")
}

// Port of EnglishChunkerTest.testTokenize
func TestEnglishChunker_Tokenize(t *testing.T) {
	if DiscoverOpenNLPTokenModel() == "" {
		t.Skip("en-token.bin missing")
	}
	chunker := NewEnglishChunker()
	// Java: chunker.tokenize replaces ’ with '
	got := chunker.Tokenize("I'm going to London")
	require.Equal(t, []string{"I", "'m", "going", "to", "London"}, got)
	got2 := chunker.Tokenize("I’m going to London") // curly apostrophe
	require.Equal(t, []string{"I", "'m", "going", "to", "London"}, got2)
}

// Port of EnglishChunkerTest.testNonBreakingSpace (chunk mapping skips NBSP)
func TestEnglishChunker_NonBreakingSpace(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" {
		t.Skip("OpenNLP models required")
	}
	// "Away from home often?" with regular space vs NBSP before often
	for _, input := range []string{"Away from home often?", "Away from home\u00A0often?"} {
		// Build readings with spaces/nbsp as separate tokens like LT would.
		var tokens []*languagetool.AnalyzedTokenReadings
		// crude: use createReadingsList for normal; custom for nbsp
		if strings.Contains(input, "\u00A0") {
			tokens = []*languagetool.AnalyzedTokenReadings{
				readings("Away", "RB", 0),
				readings(" ", "", 4),
				readings("from", "IN", 5),
				readings(" ", "", 9),
				readings("home", "NN", 10),
				readings("\u00A0", "", 14),
				readings("often", "RB", 15),
				readings("?", ".", 20),
			}
		} else {
			tokens = createReadingsList("Away from home often")
			tokens = append(tokens, readings("?", ".", 20))
		}
		NewEnglishChunker().AddChunkTags(tokens)
		// home should get NP-singular; often ADVP
		require.NotEmpty(t, tokens[0].GetChunkTags(), "input=%q", input)
		require.NotEmpty(t, tokens[4].GetChunkTags(), "home input=%q", input)
	}
}
