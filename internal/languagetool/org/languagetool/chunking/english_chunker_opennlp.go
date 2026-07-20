package chunking

import (
	"strings"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// OpenNLP-backed path for EnglishChunker — faithful to Java:
//   tokenize(sentence) → cleanZeroWidth → posTag → chunk → map exact spans → filter.
// Requires en-token.bin, en-pos-maxent.bin, and en-chunker.bin under third_party/opennlp-models.

var (
	openNLPFullOnce sync.Once
	openNLPTok      *TokenizerME
	openNLPPOS      *POSTaggerME
	openNLPChunk    *ChunkerME
	openNLPFullOK   bool
)

func getOpenNLPEnglishPipeline() (tok *TokenizerME, pos *POSTaggerME, chunk *ChunkerME, ok bool) {
	openNLPFullOnce.Do(func() {
		tp := DiscoverOpenNLPTokenModel()
		pp := DiscoverOpenNLPPOSModel()
		cp := DiscoverOpenNLPChunkerModel()
		if tp == "" || pp == "" || cp == "" {
			return
		}
		t, err := NewTokenizerME(tp)
		if err != nil || t == nil {
			return
		}
		p, err := NewPOSTaggerME(pp)
		if err != nil || p == nil {
			return
		}
		c, err := NewChunkerME(cp)
		if err != nil || c == nil {
			return
		}
		openNLPTok = t
		openNLPPOS = p
		openNLPChunk = c
		openNLPFullOK = true
	})
	return openNLPTok, openNLPPOS, openNLPChunk, openNLPFullOK
}

// tryOpenNLPChunks runs the full Java EnglishChunker OpenNLP pipeline when models load.
// Returns false when models are missing or the pipeline cannot run — incomplete, not invent.
func (c *EnglishChunker) tryOpenNLPChunks(tokens []*languagetool.AnalyzedTokenReadings) bool {
	tokME, posME, chunkME, ok := getOpenNLPEnglishPipeline()
	if !ok || tokME == nil || posME == nil || chunkME == nil || len(tokens) == 0 {
		return false
	}

	sentence := getSentenceFromReadings(tokens)
	// Java: sentence.replace('’', '\'') inside tokenize()
	openToks := tokME.Tokenize(strings.ReplaceAll(sentence, "’", "'"))
	openToks = cleanZeroWidthWhitespaces(openToks)
	if len(openToks) == 0 {
		return false
	}
	posTags := posME.Tag(openToks)
	if len(posTags) != len(openToks) {
		return false
	}
	chunkTags := chunkME.Chunk(openToks, posTags)
	if len(chunkTags) != len(openToks) {
		return false
	}

	tagged := getTokensWithTokenReadings(tokens, openToks, chunkTags)
	if c.Filter != nil {
		tagged = c.Filter.Filter(tagged)
	}
	assignChunksToReadings(tagged)
	return true
}

func getSentenceFromReadings(sentenceTokens []*languagetool.AnalyzedTokenReadings) string {
	var b strings.Builder
	for _, t := range sentenceTokens {
		if t == nil {
			continue
		}
		b.WriteString(t.GetToken())
	}
	return b.String()
}

// cleanZeroWidthWhitespaces ports EnglishChunker.cleanZeroWidthWhitespaces (including the
// quirk of re-adding the full token when a split piece is non-empty).
func cleanZeroWidthWhitespaces(tokens []string) []string {
	var clean []string
	for _, token := range tokens {
		splits := strings.Split(token, "\uFEFF")
		for _, split := range splits {
			if len(split) == 0 {
				clean = append(clean, "")
			} else {
				clean = append(clean, token)
			}
		}
	}
	return clean
}

// getTokensWithTokenReadings ports EnglishChunker.getTokensWithTokenReadings —
// OpenNLP token positions are cumulative lengths without whitespace.
func getTokensWithTokenReadings(tokenReadings []*languagetool.AnalyzedTokenReadings, tokens, chunkTags []string) []ChunkTaggedToken {
	result := make([]ChunkTaggedToken, 0, len(chunkTags))
	pos := 0
	for i, chunkTag := range chunkTags {
		startPos := pos
		endPos := startPos + len(tokens[i])
		readings := getAnalyzedTokenReadingsFor(startPos, endPos, tokenReadings)
		// Java: new ChunkTag(chunkTag) — empty throws IllegalArgumentException.
		// Do not invent "O" for missing tags (soft invent removed).
		result = append(result, NewChunkTaggedToken(tokens[i], []ChunkTag{NewChunkTag(chunkTag)}, readings))
		pos = endPos
	}
	return result
}

func assignChunksToReadings(chunkTaggedTokens []ChunkTaggedToken) {
	for _, tagged := range chunkTaggedTokens {
		if tagged.Readings == nil {
			continue
		}
		var strs []string
		for _, ct := range tagged.ChunkTags {
			if s := ct.GetChunkTag(); s != "" {
				strs = append(strs, s)
			}
		}
		tagged.Readings.SetChunkTags(strs)
	}
}

// getAnalyzedTokenReadingsFor ports EnglishChunker.getAnalyzedTokenReadingsFor —
// exact position match on non-whitespace LT tokens only.
func getAnalyzedTokenReadingsFor(startPos, endPos int, tokenReadings []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	pos := 0
	for _, tokenReading := range tokenReadings {
		if tokenReading == nil {
			continue
		}
		token := tokenReading.GetToken()
		if strings.TrimSpace(token) == "" ||
			(len(token) == 1 && unicode.IsSpace(rune(token[0]))) {
			// OpenNLP result has no whitespace tokens; skip without advancing
			// (Java: continue without pos += length).
			continue
		}
		tokenStart := pos
		tokenEnd := pos + len(token)
		if tokenStart == startPos && tokenEnd == endPos {
			return tokenReading
		}
		pos = tokenEnd
	}
	return nil
}
