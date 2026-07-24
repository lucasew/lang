package chunking

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EnglishChunker ports org.languagetool.chunking.EnglishChunker.
//
// Java always loads OpenNLP en-token / en-pos-maxent / en-chunker models and
// runs TokenizerME → POSTaggerME → ChunkerME → exact-span map → EnglishChunkFilter.
// No invent POS→BIO path: missing models leave tokens unchunked (incomplete).
type EnglishChunker struct {
	Filter *EnglishChunkFilter
}

func NewEnglishChunker() *EnglishChunker {
	return &EnglishChunker{
		Filter: NewEnglishChunkFilter(),
	}
}

// Tokenize ports package-private EnglishChunker.tokenize (OpenNLP TokenizerME).
// Replaces ’ with ' as Java does. Returns nil if en-token.bin is unavailable.
func (c *EnglishChunker) Tokenize(sentence string) []string {
	tok, _, _, ok := getOpenNLPEnglishPipeline()
	if !ok || tok == nil {
		// Lazy-load tokenizer alone for tests when only token model exists.
		p := DiscoverOpenNLPTokenModel()
		if p == "" {
			return nil
		}
		t, err := NewTokenizerME(p)
		if err != nil || t == nil {
			return nil
		}
		return t.Tokenize(strings.ReplaceAll(sentence, "’", "'"))
	}
	return tok.Tokenize(strings.ReplaceAll(sentence, "’", "'"))
}

// AddChunkTags implements Chunker (Java EnglishChunker.addChunkTags).
// OpenNLP only — same as Java; no invent POS→BIO when models are missing.
func (c *EnglishChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	c.tryOpenNLPChunks(tokens)
}

var _ Chunker = (*EnglishChunker)(nil)
