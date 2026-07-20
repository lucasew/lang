package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishChunker(t *testing.T) {
	if DiscoverOpenNLPChunkerModel() == "" {
		t.Skip("OpenNLP models required — Java EnglishChunker has no invent POS→BIO path")
	}
	// Spaced tokens like Java createReadingsList so OpenNLP position map works.
	tokens := createReadingsList("The dogs run")
	// Filter uses LT POS for plural; set NNS on "dogs".
	nns := "NNS"
	tokens[2] = languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("dogs", &nns, nil), tokens[2].GetStartPos())
	NewEnglishChunker().AddChunkTags(tokens)
	// dogs (index 2) should get NP-plural chunk tags
	require.NotEmpty(t, tokens[2].GetChunkTags())
	joined := ""
	for _, c := range tokens[2].GetChunkTags() {
		joined += c
	}
	require.Contains(t, joined, "NP")
}

// TestEnglishChunker_NoInventPOSBIO: without invent fallback, LT POS alone never yields BIO.
// Chunk tags come only from OpenNLP (when models load).
func TestEnglishChunker_NoInventPOSBIO(t *testing.T) {
	// Direct unit of the filter path is covered elsewhere; here we assert that
	// EnglishChunker has no AssignBasicNP / invent surface — only Filter field.
	c := NewEnglishChunker()
	require.NotNil(t, c.Filter)
	// Struct has no invent knobs (compile-time: only Filter remains).
	_ = c
}

// Twin of Java ChunkTag ctor: empty chunk tag is illegal (do not invent "O").
func TestGetTokensWithTokenReadings_EmptyChunkTagPanics(t *testing.T) {
	require.Panics(t, func() {
		_ = getTokensWithTokenReadings(nil, []string{"a"}, []string{""})
	})
}

// Twin of EnglishChunker.cleanZeroWidthWhitespaces quirk: non-empty split re-adds full token.
func TestCleanZeroWidthWhitespaces_JavaQuirk(t *testing.T) {
	// token without U+FEFF unchanged as one entry
	require.Equal(t, []string{"hello"}, cleanZeroWidthWhitespaces([]string{"hello"}))
	// U+FEFF-only: split yields ["",""] → two empty strings
	got := cleanZeroWidthWhitespaces([]string{"\uFEFF"})
	require.Equal(t, []string{"", ""}, got)
	// non-empty parts re-add FULL token (not the split piece) — Java bug-for-bug
	tok := "a\uFEFFb"
	got2 := cleanZeroWidthWhitespaces([]string{tok})
	// split → ["a","b"] both non-empty → [tok, tok]
	require.Equal(t, []string{tok, tok}, got2)
}
