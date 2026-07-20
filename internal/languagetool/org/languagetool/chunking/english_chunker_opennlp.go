package chunking

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// OpenNLP-backed path for EnglishChunker (Java uses ChunkerME + POS + tokenizer).
// When en-chunker.bin loads, we run ChunkerME on non-whitespace LT tokens using
// their first POS reading (Penn tags). Full OpenNLP re-tokenization + POSME is
// a follow-up; chunk model is the critical gap vs invent POS→BIO.

var (
	openNLPChunkOnce sync.Once
	openNLPChunker   *ChunkerME
	openNLPChunkOK   bool
)

func getOpenNLPChunker() *ChunkerME {
	openNLPChunkOnce.Do(func() {
		p := DiscoverOpenNLPChunkerModel()
		if p == "" {
			return
		}
		c, err := NewChunkerME(p)
		if err != nil || c == nil {
			return
		}
		openNLPChunker = c
		openNLPChunkOK = true
	})
	if openNLPChunkOK {
		return openNLPChunker
	}
	return nil
}

// tryOpenNLPChunks runs OpenNLP ChunkerME when the model is available.
// Returns false to fall back to POS→BIO interim path.
func (c *EnglishChunker) tryOpenNLPChunks(tokens []*languagetool.AnalyzedTokenReadings) bool {
	me := getOpenNLPChunker()
	if me == nil || len(tokens) == 0 {
		return false
	}
	var idxs []int
	var toks, tags []string
	for i, t := range tokens {
		if t == nil {
			continue
		}
		tok := t.GetToken()
		if tok == "" || strings.TrimSpace(tok) == "" {
			continue
		}
		pos := firstTokenPOS(t)
		if pos == "" {
			pos = "NN" // OpenNLP models expect a tag; untagged rare in full pipeline
		}
		idxs = append(idxs, i)
		toks = append(toks, tok)
		tags = append(tags, pos)
	}
	if len(toks) == 0 {
		return false
	}
	// OpenNLP tokenizer expects ASCII apostrophe
	for i, t := range toks {
		toks[i] = strings.ReplaceAll(t, "’", "'")
	}
	chunks := me.Chunk(toks, tags)
	if len(chunks) != len(toks) {
		return false
	}
	// Build ChunkTaggedToken list for EnglishChunkFilter
	tagged := make([]ChunkTaggedToken, len(toks))
	for i := range toks {
		ct := chunks[i]
		if ct == "" {
			ct = "O"
		}
		tagged[i] = NewChunkTaggedToken(toks[i], []ChunkTag{NewChunkTag(ct)}, tokens[idxs[i]])
	}
	if c.Filter != nil {
		tagged = c.Filter.Filter(tagged)
	}
	for j, t := range tagged {
		if j >= len(idxs) {
			break
		}
		i := idxs[j]
		var strs []string
		for _, ct := range t.ChunkTags {
			if s := ct.GetChunkTag(); s != "" {
				strs = append(strs, s)
			}
		}
		tokens[i].SetChunkTags(strs)
	}
	return true
}

func firstTokenPOS(t *languagetool.AnalyzedTokenReadings) string {
	if t == nil {
		return ""
	}
	for _, r := range t.GetReadings() {
		if r == nil {
			continue
		}
		p := r.GetPOSTag()
		if p == nil || *p == "" {
			continue
		}
		pos := *p
		if pos == languagetool.SentenceStartTagName || pos == languagetool.SentenceEndTagName ||
			pos == languagetool.ParagraphEndTagName {
			continue
		}
		return pos
	}
	return ""
}
