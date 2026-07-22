package chunking

import "fmt"

// EnglishChunkFilter ports org.languagetool.chunking.EnglishChunkFilter.
// Our chunker detects noun phrases but not whether they are singular
// or plural noun phrases. We add this information here.
// @since 2.3 (Java)
type EnglishChunkFilter struct{}

func NewEnglishChunkFilter() *EnglishChunkFilter { return &EnglishChunkFilter{} }

// Java: private static final ChunkTag BEGIN_NOUN_PHRASE_TAG = new ChunkTag("B-NP");
// Java: private static final ChunkTag IN_NOUN_PHRASE_TAG = new ChunkTag("I-NP");
var (
	beginNounPhraseTag = NewChunkTag("B-NP")
	inNounPhraseTag    = NewChunkTag("I-NP")
)

// chunkType ports EnglishChunkFilter.ChunkType
type chunkType int

const (
	chunkSingular chunkType = iota
	chunkPlural
)

// Filter ports EnglishChunkFilter.filter.
// Rewrites NP chunk tags with singular/plural variants.
func (f *EnglishChunkFilter) Filter(tokens []ChunkTaggedToken) []ChunkTaggedToken {
	// Java always constructs a new ArrayList; nil receiver / empty is a Go convenience.
	if f == nil {
		return tokens
	}
	result := make([]ChunkTaggedToken, 0, len(tokens))
	var newChunkTag string
	for i, taggedToken := range tokens {
		var chunkTags []ChunkTag
		if isBeginningOfNounPhrase(taggedToken) {
			ct := getChunkType(tokens, i)
			if ct == chunkSingular || endOfNounPhraseIsSingular(tokens, i) {
				chunkTags = append(chunkTags, NewChunkTag("B-NP-singular"))
				newChunkTag = "NP-singular"
			} else if ct == chunkPlural {
				chunkTags = append(chunkTags, NewChunkTag("B-NP-plural"))
				newChunkTag = "NP-plural"
			} else {
				// Java: throw new IllegalStateException("Unknown chunk type: " + chunkType);
				panic(fmt.Sprintf("Unknown chunk type: %v", ct))
			}
		}
		if newChunkTag != "" && isEndOfNounPhrase(tokens, i) {
			chunkTags = append(chunkTags, NewChunkTag("E-"+newChunkTag))
			newChunkTag = ""
		}
		if newChunkTag != "" && isContinuationOfNounPhrase(taggedToken) {
			chunkTags = append(chunkTags, NewChunkTag("I-"+newChunkTag))
		}
		if len(chunkTags) > 0 {
			result = append(result, NewChunkTaggedToken(taggedToken.Token, chunkTags, taggedToken.Readings))
		} else {
			result = append(result, taggedToken)
		}
	}
	return result
}

// endOfNounPhraseIsSingular ports EnglishChunkFilter.endOfNounPhraseIsSingular
func endOfNounPhraseIsSingular(tokens []ChunkTaggedToken, i int) bool {
	for j := i; j < len(tokens); j++ {
		if isEndOfNounPhrase(tokens, j) {
			return getChunkType(tokens, j) == chunkSingular
		}
	}
	return false
}

// isBeginningOfNounPhrase ports EnglishChunkFilter.isBeginningOfNounPhrase
func isBeginningOfNounPhrase(taggedToken ChunkTaggedToken) bool {
	// Java: taggedToken.getChunkTags().contains(BEGIN_NOUN_PHRASE_TAG)
	for _, c := range taggedToken.ChunkTags {
		if c.Equal(beginNounPhraseTag) {
			return true
		}
	}
	return false
}

// isEndOfNounPhrase ports EnglishChunkFilter.isEndOfNounPhrase
func isEndOfNounPhrase(tokens []ChunkTaggedToken, i int) bool {
	// Java: if (i > tokens.size() - 2) return true;
	if i > len(tokens)-2 {
		return true
	}
	// Java: if (!isContinuationOfNounPhrase(tokens.get(i + 1))) return true;
	if !isContinuationOfNounPhrase(tokens[i+1]) {
		return true
	}
	return false
}

// isContinuationOfNounPhrase ports EnglishChunkFilter.isContinuationOfNounPhrase
func isContinuationOfNounPhrase(taggedToken ChunkTaggedToken) bool {
	// Java: taggedToken.getChunkTags().contains(IN_NOUN_PHRASE_TAG)
	for _, c := range taggedToken.ChunkTags {
		if c.Equal(inNounPhraseTag) {
			return true
		}
	}
	return false
}

// getChunkType ports EnglishChunkFilter.getChunkType
// Get the type of the chunk that starts at the given position.
func getChunkType(tokens []ChunkTaggedToken, chunkStartPos int) chunkType {
	isPlural := false
	for i := chunkStartPos; i < len(tokens); i++ {
		token := tokens[i]
		if !isBeginningOfNounPhrase(token) && !isContinuationOfNounPhrase(token) {
			break
		}
		// Java: if (false && "and".equals(token.getToken())) {
		//   // e.g. "Tarzan and Jane" is a plural noun phrase
		//   // TODO: "Additionally, there are over 500 college and university chapter."
		//   isPlural = true;
		// } else if (hasNounWithPluralReading(token)) {
		if false && token.GetToken() == "and" {
			// disabled branch kept for 1:1 structure with Java
			isPlural = true
		} else if hasNounWithPluralReading(token) {
			// e.g. "ten books" is a plural noun phrase
			isPlural = true
		}
	}
	if isPlural {
		return chunkPlural
	}
	return chunkSingular
}

// hasNounWithPluralReading ports EnglishChunkFilter.hasNounWithPluralReading
func hasNounWithPluralReading(token ChunkTaggedToken) bool {
	if token.GetReadings() != nil {
		for _, analyzedToken := range token.GetReadings().GetReadings() {
			if analyzedToken == nil {
				continue
			}
			// Java: "NNS".equals(analyzedToken.getPOSTag())
			pt := analyzedToken.GetPOSTag()
			if pt != nil && *pt == "NNS" {
				return true
			}
		}
	}
	return false
}
