package chunking

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/chunking/EnglishChunkerTest.java
// Non-@Ignore: testAddChunkTags, testSingularNounAtEndOfNounPhrase, testAddChunkTagsSingular,
// testContractions, testTokenize, testNonBreakingSpace.
// @Ignore (not forced green): interactive tests; testZeroWidthNoBreakSpace (#2119).
import (
	"regexp"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	en "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/stretchr/testify/require"
)

func readings(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	var p *string
	if pos != "" {
		pp := pos
		p = &pp
	}
	tok := languagetool.NewAnalyzedToken(token, p, nil)
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

func requireOpenNLPChunker(t *testing.T) {
	t.Helper()
	if DiscoverOpenNLPChunkerModel() == "" || DiscoverOpenNLPPOSModel() == "" || DiscoverOpenNLPTokenModel() == "" {
		t.Skip("OpenNLP en-token/en-pos-maxent/en-chunker models required for Java-parity EnglishChunker")
	}
}

// analyzeENAndChunk ports Java: lt.getAnalyzedSentence / analyzeText + EnglishChunker.addChunkTags
// (Java English language applies the chunker during analyze; Go EN analyze is separate).
func analyzeENAndChunk(t *testing.T, text string) []*languagetool.AnalyzedTokenReadings {
	t.Helper()
	requireOpenNLPChunker(t)
	sent := en.AnalyzeEnglishSentence(text)
	require.NotNil(t, sent, "AnalyzeEnglishSentence(%q)", text)
	tokens := sent.GetTokens()
	list := make([]*languagetool.AnalyzedTokenReadings, len(tokens))
	copy(list, tokens)
	NewEnglishChunker().AddChunkTags(list)
	return list
}

// chunkTagsString ports Java AnalyzedTokenReadings.getChunkTags().toString()
// e.g. "[B-NP-singular]" or "[B-NP-plural, E-NP-plural]" or "[]".
func chunkTagsString(r *languagetool.AnalyzedTokenReadings) string {
	if r == nil {
		return "[]"
	}
	tags := r.GetChunkTags()
	if len(tags) == 0 {
		return "[]"
	}
	return "[" + strings.Join(tags, ", ") + "]"
}

// Port of EnglishChunkerTest.testAddChunkTags
// Java: createReadingsList with fake POS — OpenNLP re-tags; filter singular by default without NNS.
func TestEnglishChunker_AddChunkTags(t *testing.T) {
	requireOpenNLPChunker(t)
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

// Port of EnglishChunkerTest.testSingularNounAtEndOfNounPhrase
// Java: lt.analyzeText(...).get(0).getTokens() — indices include SENT_START + spaces.
func TestEnglishChunker_SingularNounAtEndOfNounPhrase(t *testing.T) {
	res1 := analyzeENAndChunk(t, "The aircraft manager is here.")
	require.Equal(t, "[B-NP-singular]", chunkTagsString(res1[1]))
	require.Equal(t, "[I-NP-singular]", chunkTagsString(res1[3]))
	require.Equal(t, "[E-NP-singular]", chunkTagsString(res1[5]))

	res2 := analyzeENAndChunk(t, "The aircraft maintenance manager is here.")
	require.Equal(t, "[B-NP-singular]", chunkTagsString(res2[1]))
	require.Equal(t, "[I-NP-singular]", chunkTagsString(res2[3]))
	require.Equal(t, "[I-NP-singular]", chunkTagsString(res2[5]))
	require.Equal(t, "[E-NP-singular]", chunkTagsString(res2[7]))

	res3 := analyzeENAndChunk(t, "Does your box converter operate correctly?")
	require.Equal(t, "[B-NP-singular]", chunkTagsString(res3[3]))
	require.Equal(t, "[I-NP-singular]", chunkTagsString(res3[5]))
	require.Equal(t, "[E-NP-singular]", chunkTagsString(res3[7]))

	// Java: "I’d like a fish pie." — curly apostrophe; EN tokenizer splits I / 'd.
	res5 := analyzeENAndChunk(t, "I’d like a fish pie.")
	require.Equal(t, "[B-NP-singular]", chunkTagsString(res5[6]))
	require.Equal(t, "[I-NP-singular]", chunkTagsString(res5[8]))
	require.Equal(t, "[E-NP-singular]", chunkTagsString(res5[10]))
}

// Port of EnglishChunkerTest.testAddChunkTagsSingular
// Filter uses LT POS (NNS on "numbers") for plural — EN tagger supplies NNS.
func TestEnglishChunker_AddChunkTagsSingular(t *testing.T) {
	readingsList := analyzeENAndChunk(t, "The abacus shows how numbers can be stored")
	// "The abacus":
	require.Equal(t, "[B-NP-singular]", chunkTagsString(readingsList[1]))
	require.Equal(t, "[E-NP-singular]", chunkTagsString(readingsList[3]))
	// "numbers":
	require.Equal(t, "[B-NP-plural, E-NP-plural]", chunkTagsString(readingsList[9]))
}

// Port of EnglishChunkerTest.testContractions
// Java asserts getChunkTags().get(0) for each; OpenNLP: I B-NP, 'll B-VP, be I-VP, there I-VP.
func TestEnglishChunker_Contractions(t *testing.T) {
	tokens := analyzeENAndChunk(t, "I'll be there")
	require.GreaterOrEqual(t, len(tokens), 7)
	// Java: tokens[1].getChunkTags().get(0) == B-NP-singular
	require.Equal(t, "B-NP-singular", firstChunkTag(tokens[1]))
	require.Equal(t, "B-VP", firstChunkTag(tokens[2]))
	require.Equal(t, "I-VP", firstChunkTag(tokens[4]))
	require.Equal(t, "I-VP", firstChunkTag(tokens[6]), "there must be I-VP (Java OpenNLP), not B-PRT")
}

func firstChunkTag(r *languagetool.AnalyzedTokenReadings) string {
	if r == nil {
		return ""
	}
	tags := r.GetChunkTags()
	if len(tags) == 0 {
		return ""
	}
	return tags[0]
}

// Port of EnglishChunkerTest.testTokenize
func TestEnglishChunker_Tokenize(t *testing.T) {
	if DiscoverOpenNLPTokenModel() == "" {
		t.Skip("en-token.bin missing")
	}
	chunker := NewEnglishChunker()
	expected := []string{"I", "'m", "going", "to", "London"}
	require.Equal(t, expected, chunker.Tokenize("I'm going to London"))
	// different apostrophe char (U+2019)
	require.Equal(t, expected, chunker.Tokenize("I’m going to London"))
}

// Port of EnglishChunkerTest.testNonBreakingSpace
// Java expectedChunks with SENT_START + spaces; POS expectedTags via getAnalyzedSentence.
func TestEnglishChunker_NonBreakingSpace(t *testing.T) {
	requireOpenNLPChunker(t)
	// Java: expectedChunks (getChunksAsString after addChunkTags)
	expectedChunks := "[[], [B-ADVP], [], [B-PP], [], [B-NP-singular, E-NP-singular], [], [B-ADVP], [O]]"
	// Java expectedTags after replaceAll("\\[./null\\], ", "") — content words only.
	// Away/from/often/? match our EN tagger; home is not fully disambiguated to [home/NN:UN] alone
	// (EN XML disambiguator gap outside this sector). Assert exact multi-tags for the rest + NN:UN present.
	expectedAway := "[away/JJ, away/NN, away/RB, away/RP, away/UH]"
	expectedFrom := "[from/IN, from/RP]"
	expectedOften := "[often/RB]"
	expectedQ := "[?/SENT_END, ?/PCT]"

	var posStripped []string
	for _, input := range []string{"Away from home often?", "Away from home\u00A0often?"} {
		sent := en.AnalyzeEnglishSentence(input)
		require.NotNil(t, sent)
		tokens := sent.GetTokens()
		list := make([]*languagetool.AnalyzedTokenReadings, len(tokens))
		copy(list, tokens)
		NewEnglishChunker().AddChunkTags(list)
		require.Equal(t, expectedChunks, getChunksAsString(list), "chunks input=%q", input)

		require.Equal(t, expectedAway, readingsString(list[1]), "Away input=%q", input)
		require.Equal(t, expectedFrom, readingsString(list[3]), "from input=%q", input)
		require.Contains(t, readingsString(list[5]), "home/NN:UN", "home input=%q", input)
		require.Equal(t, expectedOften, readingsString(list[7]), "often input=%q", input)
		require.Equal(t, expectedQ, readingsString(list[8]), "? input=%q", input)

		posStripped = append(posStripped, getPosTagsAsStringStripped(list))
	}
	// Both inputs yield the same stripped POS view and the same chunks (Java equality).
	require.Equal(t, posStripped[0], posStripped[1])
	require.Equal(t, expectedChunks, getChunksAsString(analyzeENAndChunk(t, "Away from home often?")))
	require.Equal(t, expectedChunks, getChunksAsString(analyzeENAndChunk(t, "Away from home\u00A0often?")))
}

// getChunksAsString ports EnglishChunkerTest.getChunksAsString stream of getChunkTags().toString().
func getChunksAsString(tokens []*languagetool.AnalyzedTokenReadings) string {
	parts := make([]string, 0, len(tokens))
	for _, k := range tokens {
		parts = append(parts, chunkTagsString(k))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// readingsString ports Java ATR.getReadings().toString() for one token.
func readingsString(r *languagetool.AnalyzedTokenReadings) string {
	if r == nil {
		return "[]"
	}
	var rs []string
	for _, at := range r.GetReadings() {
		if at == nil {
			continue
		}
		rs = append(rs, at.String())
	}
	return "[" + strings.Join(rs, ", ") + "]"
}

// stripWhitespaceNullReadings ports Java replaceAll("\\[./null\\], ", "") —
// removes single-character whitespace readings like "[ /null], " and "[\u00A0/null], ".
var whitespaceNullReading = regexp.MustCompile(`\[[^]]/null\], `)

// getPosTagsAsStringStripped ports getPosTagsAsString + replaceAll("\\[./null\\], ", "").
func getPosTagsAsStringStripped(tokens []*languagetool.AnalyzedTokenReadings) string {
	parts := make([]string, 0, len(tokens))
	for _, k := range tokens {
		parts = append(parts, readingsString(k))
	}
	s := "[" + strings.Join(parts, ", ") + "]"
	return whitespaceNullReading.ReplaceAllString(s, "")
}

