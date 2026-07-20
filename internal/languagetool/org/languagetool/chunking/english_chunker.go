package chunking

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EnglishChunker ports org.languagetool.chunking.EnglishChunker.
//
// When third_party/opennlp-models/en-chunker.bin is present, runs OpenNLP
// ChunkerME (GIS maxent + beam) on non-whitespace LT tokens using their first
// POS reading, then EnglishChunkFilter. Java also re-tokenizes and re-POS-tags
// with OpenNLP models — still incomplete vs full Java when only the chunk
// model is wired (no invent; POS→BIO fallback if the model is missing).
//
// Fallback without model:
//  1) map first LT POS → phrase type → B-/I- BIO;
//  2) EnglishChunkFilter (singular/plural NP).
type EnglishChunker struct {
	Filter *EnglishChunkFilter
	// AssignBasicNP enables POS→phrase BIO when AssignBasicNP is true (default).
	// Name retained for tests; covers NP/VP/PP/ADVP/PRT, not only NP.
	AssignBasicNP bool
	// IsNounish reports whether a POS tag is noun-like (default: NN*).
	IsNounish func(posTag string) bool
}

func NewEnglishChunker() *EnglishChunker {
	return &EnglishChunker{
		Filter:        NewEnglishChunkFilter(),
		AssignBasicNP: true,
		IsNounish: func(pos string) bool {
			return len(pos) >= 2 && pos[0] == 'N' && pos[1] == 'N'
		},
	}
}

// AddChunkTags implements Chunker (Java EnglishChunker.addChunkTags).
// Java OpenNLP chunker runs on non-whitespace tokens only.
func (c *EnglishChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	// Prefer OpenNLP ChunkerME when en-chunker.bin is available (Java path).
	if c.tryOpenNLPChunks(tokens) {
		return
	}
	var idxs []int
	var tagged []ChunkTaggedToken
	for i, t := range tokens {
		if t == nil {
			continue
		}
		tok := t.GetToken()
		if tok != "" && strings.TrimSpace(tok) == "" {
			continue
		}
		if tok == "" {
			// SENT_START / empty: omit from phrase stream
			continue
		}
		var tags []ChunkTag
		for _, ct := range t.GetChunkTags() {
			tags = append(tags, NewChunkTag(ct))
		}
		idxs = append(idxs, i)
		tagged = append(tagged, NewChunkTaggedToken(tok, tags, t))
	}
	if c.AssignBasicNP {
		tagged = c.assignPOSBasedBIO(tagged)
	}
	if c.Filter != nil {
		tagged = c.Filter.Filter(tagged)
	}
	for j, t := range tagged {
		if j >= len(idxs) {
			break
		}
		i := idxs[j]
		if i >= len(tokens) || tokens[i] == nil {
			continue
		}
		var strs []string
		for _, ct := range t.ChunkTags {
			s := ct.GetChunkTag()
			// Keep "O" so chunk_re="…|O" matches (Java OpenNLP outside tag).
			if s != "" {
				strs = append(strs, s)
			}
		}
		tokens[i].SetChunkTags(strs)
	}
}

// assignPOSBasedBIO maps first POS → phrase type → B-/I- tags.
// No surface-list invent; multi-tag disambiguation invent removed (use first reading).
func (c *EnglishChunker) assignPOSBasedBIO(tokens []ChunkTaggedToken) []ChunkTaggedToken {
	out := make([]ChunkTaggedToken, len(tokens))
	copy(out, tokens)
	phrases := make([]string, len(tokens))
	poss := make([]string, len(tokens))
	for i, t := range out {
		pos := firstChunkTokenPOS(t)
		poss[i] = pos
		phrases[i] = phraseFromPOS(pos)
	}
	bio := toBIOWithPOS(phrases, poss)
	for i := range out {
		if bio[i] == "" || bio[i] == "O" {
			if len(out[i].ChunkTags) == 0 {
				out[i].ChunkTags = []ChunkTag{NewChunkTag("O")}
			}
			continue
		}
		out[i].ChunkTags = []ChunkTag{NewChunkTag(bio[i])}
	}
	return out
}

// firstChunkTokenPOS returns the first non-boundary POS reading (Java OpenNLP gets one POS).
func firstChunkTokenPOS(t ChunkTaggedToken) string {
	if t.Readings == nil {
		return ""
	}
	for _, r := range t.Readings.GetReadings() {
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

func phraseFromPOS(pos string) string {
	switch {
	case pos == "" || pos == "," || pos == "." || strings.HasPrefix(pos, "PCT"):
		return "O"
	case strings.HasPrefix(pos, "VB") || pos == "MD":
		return "VP"
	case strings.HasPrefix(pos, "RB") || pos == "WRB":
		return "ADVP"
	case pos == "RP":
		return "PRT"
	case pos == "IN" || pos == "TO":
		return "PP"
	case strings.HasPrefix(pos, "NN") || pos == "DT" || pos == "PDT" ||
		pos == "PRP" || pos == "PRP$" || pos == "CD" || pos == "EX" ||
		pos == "WP" || pos == "WP$" || pos == "WDT" || pos == "POS" ||
		strings.HasPrefix(pos, "JJ"):
		return "NP"
	case pos == "CC":
		return "O"
	default:
		return "O"
	}
}

// toBIOWithPOS restarts NP at DT/PDT/PRP (common OpenNLP-style NP break).
func toBIOWithPOS(phrase []string, poss []string) []string {
	out := make([]string, len(phrase))
	prev := ""
	for i, p := range phrase {
		if p == "O" || p == "" {
			out[i] = "O"
			prev = ""
			continue
		}
		restart := false
		if p == "NP" && prev == "NP" && i < len(poss) {
			pos := poss[i]
			if pos == "DT" || pos == "PDT" || pos == "PRP" {
				restart = true
			}
		}
		if p == prev && !restart {
			out[i] = "I-" + p
		} else {
			out[i] = "B-" + p
		}
		prev = p
	}
	return out
}

var _ Chunker = (*EnglishChunker)(nil)
