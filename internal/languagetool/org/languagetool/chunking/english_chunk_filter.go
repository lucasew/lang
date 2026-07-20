package chunking

// EnglishChunkFilter ports org.languagetool.chunking.EnglishChunkFilter.
// Adds singular/plural NP tags to B-NP/I-NP sequences (Java twin).
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

// isEndOfNounPhrase ports Java: true when next is not I-NP (or end of list).
func isEndOfNounPhrase(tokens []ChunkTaggedToken, i int) bool {
	// Java: if (i > tokens.size() - 2) return true;
	if i > len(tokens)-2 {
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

// getChunkType ports Java EnglishChunkFilter.getChunkType.
func getChunkType(tokens []ChunkTaggedToken, i int) chunkType {
	plural := false
	for j := i; j < len(tokens); j++ {
		tok := tokens[j]
		if !isBeginningOfNounPhrase(tok) && !isContinuationOfNounPhrase(tok) {
			break
		}
		// Java "and" plural branch is disabled (if (false && "and".equals...)).
		if hasNounWithPluralReading(tok) {
			plural = true
		}
	}
	if plural {
		return chunkPlural
	}
	return chunkSingular
}

// hasNounWithPluralReading ports Java hasNounWithPluralReading.
func hasNounWithPluralReading(t ChunkTaggedToken) bool {
	if t.Readings == nil {
		return false
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
		// Java EnglishChunkFilter: "NNS".equals(analyzedToken.getPOSTag())
		if tag == "NNS" {
			return true
		}
	}
	return false
}
