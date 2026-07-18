package chunking

import (
	"strings"
)

// EnglishChunkFilter ports org.languagetool.chunking.EnglishChunkFilter.
// Adds singular/plural NP tags to B-NP/I-NP sequences.
type EnglishChunkFilter struct{}

func NewEnglishChunkFilter() *EnglishChunkFilter { return &EnglishChunkFilter{} }

type chunkType int

const (
	chunkSingular chunkType = iota
	chunkPlural
)

// Filter rewrites NP chunk tags with singular/plural variants.
func (f *EnglishChunkFilter) Filter(tokens []ChunkTaggedToken) []ChunkTaggedToken {
	if f == nil || len(tokens) == 0 {
		return tokens
	}
	result := make([]ChunkTaggedToken, 0, len(tokens))
	var newChunkTag string
	for i, tagged := range tokens {
		var chunkTags []ChunkTag
		if isBeginningOfNounPhrase(tagged) {
			ct := getChunkType(tokens, i)
			// anyone/someone: always singular NP (Java OpenNLP; not pluralized by
			// a following NNS|VBZ "knows").
			if isSingularPronounSurface(tagged.Token) {
				ct = chunkSingular
			}
			if ct == chunkSingular || endOfNounPhraseIsSingular(tokens, i) {
				chunkTags = append(chunkTags, NewChunkTag("B-NP-singular"))
				newChunkTag = "NP-singular"
			} else {
				chunkTags = append(chunkTags, NewChunkTag("B-NP-plural"))
				newChunkTag = "NP-plural"
			}
		}
		if newChunkTag != "" && isEndOfNounPhrase(tokens, i) {
			chunkTags = append(chunkTags, NewChunkTag("E-"+newChunkTag))
			newChunkTag = ""
		}
		if newChunkTag != "" && isContinuationOfNounPhrase(tagged) {
			chunkTags = append(chunkTags, NewChunkTag("I-"+newChunkTag))
		}
		if len(chunkTags) > 0 {
			result = append(result, NewChunkTaggedToken(tagged.Token, chunkTags, tagged.Readings))
		} else {
			result = append(result, tagged)
		}
	}
	return result
}

func isBeginningOfNounPhrase(t ChunkTaggedToken) bool {
	for _, c := range t.ChunkTags {
		if c.GetChunkTag() == "B-NP" {
			return true
		}
	}
	return false
}

func isContinuationOfNounPhrase(t ChunkTaggedToken) bool {
	for _, c := range t.ChunkTags {
		if c.GetChunkTag() == "I-NP" {
			return true
		}
	}
	return false
}

func isEndOfNounPhrase(tokens []ChunkTaggedToken, i int) bool {
	// Java EnglishChunkFilter.isEndOfNounPhrase: true when next is not I-NP
	// (including when next is B-NP starting a new phrase, or O/VP/…).
	if i+1 >= len(tokens) {
		return true
	}
	return !isContinuationOfNounPhrase(tokens[i+1])
}

func endOfNounPhraseIsSingular(tokens []ChunkTaggedToken, i int) bool {
	for j := i; j < len(tokens); j++ {
		if isEndOfNounPhrase(tokens, j) {
			return getChunkType(tokens, j) == chunkSingular
		}
	}
	return false
}

func getChunkType(tokens []ChunkTaggedToken, i int) chunkType {
	// Java EnglishChunkFilter.getChunkType: only scan B-NP/I-NP tokens in the
	// span; break before plural-check when the token is outside the NP (so a
	// following NNS|VBZ verb like "knows" does not make "anyone" plural).
	// Also: pure PRP (anyone/someone) is singular even if a later NNS is mis-tagged.
	plural := false
	for j := i; j < len(tokens); j++ {
		tok := tokens[j]
		if !isBeginningOfNounPhrase(tok) && !isContinuationOfNounPhrase(tok) {
			break
		}
		if isPluralToken(tok) {
			plural = true
		}
	}
	if plural {
		return chunkPlural
	}
	return chunkSingular
}

// isSingularPronounSurface: indefinite singular pronouns only.
// Java EnglishChunkFilter has no such list — getChunkType breaks at non-NP so
// "anyone knows" stays singular when knows is VP. We still force singular for
// these surfaces when a following NNS|VBZ is mis-chunked as I-NP.
// Do NOT include each/either/neither: DT_TIMES_NNS needs E-NP-plural on
// "Each times" so disambig can strip VBZ for EVERY_EACH_SINGULAR.
func isSingularPronounSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "anyone", "anybody", "someone", "somebody", "everyone", "everybody",
		"noone", "nobody", "one":
		return true
	default:
		return false
	}
}

func isPluralToken(t ChunkTaggedToken) bool {
	if t.Readings == nil {
		// heuristic: ends with s
		tok := t.Token
		return strings.HasSuffix(strings.ToLower(tok), "s") && !strings.HasSuffix(strings.ToLower(tok), "ss")
	}
	for _, r := range t.Readings.GetReadings() {
		if r == nil {
			continue
		}
		pt := r.GetPOSTag()
		if pt == nil {
			continue
		}
		tag := *pt
		if tag == "NNS" || tag == "NNPS" || strings.HasPrefix(tag, "NNS") || strings.HasPrefix(tag, "NNPS") {
			return true
		}
	}
	return false
}

// Ensure unused import for languagetool if readings used
